package server

import (
	"bytes"
	"github.com/panjf2000/gnet"
	"gogs/impl/ecs"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/serverbound"
)

func (s *Server) handleClientSettings(conn gnet.Conn, pkt pk.Packet) (out []byte, err error) {
	logger.Printf("Received client settings")
	in := serverbound.ClientSettings{}
	if err = in.FromPacket(pkt); err != nil {
		return
	}

	player := s.playerFromConn(conn)

	player.ChatMode = uint8(in.ChatMode)
	player.ViewDistance = byte(in.ViewDistance)
	if player.ViewDistance > s.ViewDistance {
		player.ViewDistance = s.ViewDistance
	}

	if !player.Online {
		player.Online = true
		out = s.initialGameData(player)
	}

	return
}

// initial game data returns all the packets that a client typically needs when joining the game encoded
func (s *Server) initialGameData(player *ecs.Player) []byte {
	buf := bytes.Buffer{}

	buf.Write(clientbound.HeldItemChange{
		Slot: pk.Byte(player.HeldItem),
	}.CreatePacket().Encode())

	buf.Write(clientbound.DeclareRecipes{
		NumRecipes: 0,
		Recipes:    nil,
	}.CreatePacket().Encode())

	buf.Write(clientbound.VanillaTags().CreatePacket().Encode())

	buf.Write(clientbound.PlayerPositionAndLook{
		X:          pk.Double(player.X),
		Y:          pk.Double(player.Y),
		Z:          pk.Double(player.Z),
		Yaw:        pk.Float(player.Yaw),
		Pitch:      pk.Float(player.Pitch),
		Flags:      0,
		TeleportID: 0,
	}.CreatePacket().Encode())

	// send player info (tab list)
	s.playerMapMutex.RLock()
	numPlayers := len(s.playerMap.uuidToPlayer)
	playerInfoArr := make([]pk.Encodable, 0, numPlayers)
	for _, p := range s.playerMap.uuidToPlayer {
		playerInfoArr = append(playerInfoArr, clientbound.PlayerInfoAddPlayer{
			UUID:           pk.UUID(p.UUID),
			Name:           pk.String(p.Name),
			NumProperties:  0,
			Properties:     nil,
			Gamemode:       pk.VarInt(p.GameMode),
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

	buf.Write(clientbound.PlayerPositionAndLook{
		X:          pk.Double(player.X),
		Y:          pk.Double(player.Y),
		Z:          pk.Double(player.Z),
		Yaw:        pk.Float(player.Yaw),
		Pitch:      pk.Float(player.Pitch),
		Flags:      0,
		TeleportID: 0,
	}.CreatePacket().Encode())

	// send inventory
	player.InventoryLock.RLock()
	buf.Write(clientbound.WindowItems{
		WindowID: 0,
		Count:    pk.Short(len(player.Inventory)),
		SlotData: player.Inventory,
	}.CreatePacket().Encode())
	player.InventoryLock.RUnlock()

	// send time update with negative time to keep sun in position
	buf.Write(clientbound.TimeUpdate{WorldAge: 0, TimeOfDay: -6000}.CreatePacket().Encode())

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

	return buf.Bytes()
}

func (s *Server) chunkDataPackets(player *ecs.Player) []byte {
	buf := bytes.Buffer{}

	biomes := make([]pk.VarInt, 1024)
	motionBlocking := make([]int64, 37)

	chunkX := int(player.X) >> 4
	chunkZ := int(player.Z) >> 4

	viewDistance := int(player.ViewDistance)

	for x := -viewDistance; x <= viewDistance; x++ {
		for z := -viewDistance; z <= viewDistance; z++ {
			// TODO: track chunk also needs chunks to be unloaded ...
			//if player.TrackChunk(x+chunkX, z+chunkZ) {
			//	continue
			//}

			column := s.world.Column(x+chunkX, z+chunkZ)

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

			blockEntities := make([]pk.NBT, len(column.BlockEntities))
			for i, v := range column.BlockEntities {
				blockEntities[i] = pk.NBT{V: v}
			}

			chunk := clientbound.ChunkData{
				ChunkX:         pk.Int(x + chunkX),
				ChunkZ:         pk.Int(z + chunkZ),
				FullChunk:      true,
				PrimaryBitMask: pk.VarInt(bitMask),
				Heightmaps: pk.NBT{
					V: clientbound.Heightmap{
						MotionBlocking: motionBlocking,
						//WorldSurface:   motion_blocking,
					},
				},
				BiomesLength:     pk.VarInt(len(biomes)),
				Biomes:           biomes,
				Size:             pk.VarInt(len(chunkDataArray.Encode())),
				Data:             chunkDataArray,
				NumBlockEntities: pk.VarInt(len(blockEntities)),
				BlockEntities:    blockEntities,
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
