package server

import (
	"bytes"
	"fmt"
	"github.com/panjf2000/gnet"
	"gogs/api/events"
	"gogs/impl/ecs"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/packetids"
)

func (s *Server) handleLoginStart(conn gnet.Conn, pkt pk.Packet) (out []byte, err error) {
	ctx := conn.Context().(connectionContext)

	var name pk.String
	if err = pkt.Unmarshal(&name); err != nil {
		return
	}

	logger.Printf("Received login start packet from player %v", name)

	// TODO: implement encryption (if online mode, send encryption request instead of following)

	//trigger login event
	event := events.PlayerLoginData{
		Name: string(name),
		Conn: conn, // for ip address bans? consider changing to just ip
	}
	events.PlayerLoginEvent.Trigger(&event)

	if event.Result == events.LoginAllowed {
		buf := bytes.Buffer{}
		u := pk.NameToUUID(string(name)) // todo: get uuid from mojang servers
		// send login success
		buf.Write(pk.Marshal(
			packetids.LoginSuccess,
			pk.UUID(u),
			name,
		).Encode())

		player := s.createPlayer(string(name), u, conn)
		buf.Write(s.joinGamePacket(player).Encode())

		buf.Write(clientbound.HeldItemChange{
			Slot: pk.Byte(player.HeldItem),
		}.CreatePacket().Encode())

		buf.Write(clientbound.DeclareRecipes{
			NumRecipes: 0,
			Recipes:    nil,
		}.CreatePacket().Encode())

		buf.Write(clientbound.VanillaTags().CreatePacket().Encode())

		buf.Write((&clientbound.PlayerPositionAndLook{}).FromPlayer(*player).CreatePacket().Encode())

		buf.Write(clientbound.UpdateViewPosition{
			ChunkX: pk.VarInt(int32(player.X) >> 4),
			ChunkZ: pk.VarInt(int32(player.Z) >> 4),
		}.CreatePacket().Encode())

		buf.Write(s.chunkDataPackets(player))

		buf.Write(clientbound.SpawnPosition{Location: pk.Position{
			X: int32(player.SpawnPosition.X),
			Y: int32(player.SpawnPosition.Y),
			Z: int32(player.SpawnPosition.Z),
		}}.CreatePacket().Encode())

		buf.Write((&clientbound.PlayerPositionAndLook{}).FromPlayer(*player).CreatePacket().Encode())

		// send inventory
		buf.Write(clientbound.WindowItems{
			WindowID: 0,
			Count:    pk.Short(len(player.Inventory)),
			SlotData: player.Inventory,
		}.CreatePacket().Encode())

		// send time update with negative time to keep sun in position
		buf.Write(clientbound.TimeUpdate{WorldAge: 0, TimeOfDay: -6000}.CreatePacket().Encode())

		s.playerMapMutex.RLock()
		numPlayers := len(s.playerMap.uuidToPlayer)
		playerInfoArr := make([]pk.Encodable, 0, numPlayers)
		for _, p := range s.playerMap.uuidToPlayer {
			playerInfoArr = append(playerInfoArr, clientbound.PlayerInfoAddPlayer{
				UUID:           pk.UUID(p.UUID),
				Name:           pk.String(p.Name),
				NumProperties:  0,
				Properties:     nil,
				Gamemode:       0,
				Ping:           1,
				HasDisplayName: false,
				DisplayName:    "",
			})
		}
		buf.Write(clientbound.PlayerInfo{
			Action:     0,
			NumPlayers: pk.VarInt(numPlayers),
			Players:    playerInfoArr,
		}.CreatePacket().Encode())

		// also add spawn player packets for players already online
		// TODO: this logic should be done elsewhere (when players enter range) (tick?)
		for _, p := range s.playerMap.uuidToPlayer {
			if p.UUID != player.UUID {
				buf.Write(clientbound.SpawnPlayer{
					EntityID:   pk.VarInt(p.ID()),
					PlayerUUID: pk.UUID(p.UUID),
					X:          pk.Double(p.X),
					Y:          pk.Double(p.Y),
					Z:          pk.Double(p.Z),
					Yaw:        pk.Angle(p.Yaw),
					Pitch:      pk.Angle(p.Pitch),
				}.CreatePacket().Encode())
			}
		}
		s.playerMapMutex.RUnlock()

		out = buf.Bytes()

		/*
			event := events.PlayerJoinData{
				Player:  api.Player(player),
				Message: fmt.Sprintf("%v has joined the game", player.Name()),
			}
			events.PlayerJoinEvent.Trigger(&event)
		*/
		s.Broadcast(fmt.Sprintf("%v has joined the game", player.Name))

		// send out player info to players online
		playerInfoPacket := clientbound.PlayerInfo{
			Action:     0,
			NumPlayers: 1,
			Players: []pk.Encodable{
				clientbound.PlayerInfoAddPlayer{
					UUID:           pk.UUID(player.UUID),
					Name:           pk.String(player.Name),
					NumProperties:  0,
					Properties:     nil,
					Gamemode:       0,
					Ping:           0,
					HasDisplayName: false,
					DisplayName:    "",
				},
			},
		}.CreatePacket()
		// TODO: spawn player should be occurring when players enter range (not join game), do logic elsewhere (tick?)
		spawnPlayerPacket := clientbound.SpawnPlayer{
			EntityID:   pk.VarInt(player.ID()),
			PlayerUUID: pk.UUID(player.UUID),
			X:          pk.Double(player.X),
			Y:          pk.Double(player.Y),
			Z:          pk.Double(player.Z),
			Yaw:        pk.Angle(player.Yaw),
			Pitch:      pk.Angle(player.Pitch),
		}.CreatePacket()

		s.broadcastPacket(playerInfoPacket, conn)
		s.broadcastPacket(spawnPlayerPacket, conn)
	} else {
		// TODO: Send disconnect packet with reason
		err = fmt.Errorf("login not allowed not yet implemented")
		_ = conn.Close()
		return
	}

	conn.SetContext(connectionContext{
		State:           playState,
		ProtocolVersion: ctx.ProtocolVersion,
	})
	return
}

func (s *Server) joinGamePacket(player *ecs.Player) pk.Packet {
	return clientbound.JoinGame{
		EntityID:     pk.Int(player.ID()),
		IsHardcore:   false,
		Gamemode:     pk.UByte(player.GameMode),
		PrevGamemode: -1,
		WorldCount:   1,
		WorldNames:   []pk.Identifier{"world"},
		DimensionCodec: pk.NBT{
			V: clientbound.DimensionCodec{
				DimensionTypes: clientbound.DimensionTypeRegistry{
					Type: "minecraft:dimension_type",
					Value: []clientbound.DimensionTypeRegistryEntry{
						{"minecraft:overworld",
							0,
							clientbound.MinecraftOverworld,
						},
					},
				},
				BiomeRegistry: clientbound.BiomeRegistry{
					Type: "minecraft:worldgen/biome",
					Value: []clientbound.BiomeRegistryEntry{
						{
							Name: "minecraft:plains",
							ID:   1,
							Element: clientbound.BiomeProperties{
								Precipitation: "none",
								Depth:         0.125,
								Temperature:   0.8,
								Scale:         0.05,
								Downfall:      0.4,
								Category:      "plains",
								Effects: clientbound.BiomeEffects{
									SkyColor:      0xFF851B,
									WaterFogColor: 329011,
									FogColor:      12638463,
									WaterColor:    4159204,
								},
							},
						},
					},
				},
			}},
		Dimension:    pk.NBT{V: clientbound.MinecraftOverworld},
		WorldName:    "world",
		HashedSeed:   0,
		MaxPlayers:   20,
		ViewDistance: 10,
		RDI:          false,
		ERS:          false,
		IsDebug:      false,
		IsFlat:       true,
	}.CreatePacket()
}

func (s *Server) chunkDataPackets(player *ecs.Player) []byte {
	// TODO: get chunks & biomes from server & based on player position
	buf := bytes.Buffer{}

	biomes := make([]pk.VarInt, 1024, 1024)
	for i := range biomes {
		biomes[i] = 1
	}

	chunkX := int(player.X) >> 4
	chunkZ := int(player.Z) >> 4

	viewDistance := int(player.ViewDistance)
	if viewDistance == 0 {
		viewDistance = 10
	}
	for x := -viewDistance; x <= viewDistance; x++ {
		for z := -viewDistance; z <= viewDistance; z++ {
			column := s.world.GetColumn(x+chunkX, z+chunkZ)

			var chunkDataArray clientbound.ChunkDataArray
			chunkDataArray = make(clientbound.ChunkDataArray, 0, 16)

			bitMask := 0
			for _, section := range column.Sections {
				if section == nil {
					continue
				}
				bitMask |= 1 << section.Y

				palette := make([]pk.VarInt, len(section.Palette))
				for i, blockID := range section.Palette {
					palette[i] = pk.VarInt(blockID)
				}

				blockData := make([]pk.Long, len(section.BlockStates.Data))
				for i, blockState := range section.BlockStates.Data {
					blockData[i] = pk.Long(blockState)
				}
				chunkDataArray = append(chunkDataArray, clientbound.ChunkSection{
					BlockCount:   4096,
					BitsPerBlock: pk.UByte(section.BlockStates.BitsPerValue),
					Palette: clientbound.ChunkPalette{
						Length:  pk.VarInt(len(palette)),
						Palette: palette,
					},
					DataArrayLength: pk.VarInt(len(blockData)),
					DataArray:       blockData,
				})
			}

			chunk := clientbound.ChunkData{
				ChunkX:         pk.Int(x + chunkX),
				ChunkZ:         pk.Int(z + chunkZ),
				FullChunk:      true,
				PrimaryBitMask: pk.VarInt(bitMask),
				Heightmaps: pk.NBT{
					V: clientbound.Heightmap{
						MotionBlocking: make([]int64, 37),
						WorldSurface:   make([]int64, 37),
					},
				},
				BiomesLength:     1024,
				Biomes:           biomes,
				Size:             pk.VarInt(len(chunkDataArray.Encode())),
				Data:             chunkDataArray,
				NumBlockEntities: 0,
				BlockEntities:    nil,
			}.CreatePacket().Encode()
			buf.Write(chunk)

			//temp := make([]byte, 2048)
			//for i := range temp {
			//	temp[i] = 255
			//}
			//
			//updateLight := clientbound.UpdateLight{
			//	ChunkX:         pk.VarInt(x + chunkX),
			//	ChunkZ:         pk.VarInt(z + chunkZ),
			//	TrustEdges:          false,
			//	SkyLightMask:        0,
			//	BlockLightMask:      1 << 5,
			//	EmptySkyLightMask:   0,
			//	EmptyBlockLightMask: 0,
			//	SkyLightArrays:      clientbound.SkyLight{
			//	},
			//	BlockLightArrays:    clientbound.BlockLight{
			//		Arrays: []pk.ByteArray{
			//			temp,
			//		},
			//	},
			//}
			//buf.Write(updateLight.CreatePacket().Encode())
		}
	}
	return buf.Bytes()
}
