package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/panjf2000/gnet"
	"gogs/api"
	"gogs/api/events"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/packetids"
)

func LoginStart(c gnet.Conn, pkt *pk.Packet, s api.Server) (out []byte, err error) {
	if pkt.ID != 0 {
		return nil, errors.New("login start expects Packet ID 0")
	}

	var name pk.String
	if err = pkt.Unmarshal(&name); err != nil {
		return nil, err
	}

	logger.Printf("Received login start packet from player %v", name)

	//trigger login event
	event := events.PlayerLoginData{
		Name: string(name),
		Conn: c, // for ip address bans? consider changing to just ip
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

		// send join game and a bunch of other things
		player := s.CreatePlayer(string(name), u, c)
		buf.Write(clientbound.JoinGame{
			EntityID:     12193,
			IsHardcore:   false,
			Gamemode:     0,
			PrevGamemode: 0,
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
			IsFlat:       false,
		}.CreatePacket().Encode())

		buf.Write(clientbound.HeldItemChange{}.CreatePacket().Encode())

		buf.Write(clientbound.DeclareRecipes{
			NumRecipes: 0,
			Recipes:    nil,
		}.CreatePacket().Encode())

		buf.Write((&clientbound.PlayerPositionAndLook{}).FromPlayer(player).CreatePacket().Encode())

		buf.Write(clientbound.UpdateViewPosition{
			ChunkX: 0,
			ChunkZ: 0,
		}.CreatePacket().Encode())

		// TODO: get chunks & biomes from server
		biomes := make([]pk.VarInt, 1024, 1024)
		for i := range biomes {
			biomes[i] = 1
		}

		blockData := make([]pk.Long, 256)
		for i := 0; i < 16; i++ {
			blockData[i] = 0x1111111111111111
		}

		for x := -6; x < 6; x++ {
			for z := -6; z < 6; z++ {
				chunk := clientbound.ChunkData{
					ChunkX:         pk.Int(x),
					ChunkZ:         pk.Int(z),
					FullChunk:      true,
					PrimaryBitMask: 1,
					Heightmaps: pk.NBT{
						V: clientbound.Heightmap{
							MotionBlocking: make([]int64, 37),
							WorldSurface:   make([]int64, 37),
						},
					},
					BiomesLength: 1024,
					Biomes:       biomes,
					Size:         2056,
					Data: clientbound.ChunkDataArray{
						clientbound.ChunkSection{
							BlockCount:   64,
							BitsPerBlock: 4,
							Palette: clientbound.ChunkPalette{
								Length:  2,
								Palette: []pk.VarInt{0, 1},
							},
							DataArrayLength: 256,
							DataArray:       blockData,
						},
					},
					NumBlockEntities: 0,
					BlockEntities:    nil,
				}.CreatePacket().Encode()
				buf.Write(chunk)
			}
		}

		buf.Write(clientbound.SpawnPosition{Location: pk.Position{
			X: 0,
			Y: 2,
			Z: 0,
		}}.CreatePacket().Encode())

		buf.Write((&clientbound.PlayerPositionAndLook{}).FromPlayer(player).CreatePacket().Encode())

		// send time update with negative time to keep sun in position
		buf.Write(clientbound.TimeUpdate{WorldAge: 0, TimeOfDay: -6000}.CreatePacket().Encode())

		// send list of players who are online
		players := s.Players()
		playerInfoArr := make([]pk.Encodable, 0, len(players))
		for _, p := range players {
			playerInfoArr = append(playerInfoArr, clientbound.PlayerInfoAddPlayer{
				UUID:           pk.UUID(p.UUID()),
				Name:           pk.String(p.Name()),
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
			NumPlayers: pk.VarInt(len(players)),
			Players:    playerInfoArr,
		}.CreatePacket().Encode())

		// also add spawn player packets for players already online
		// TODO: this logic should be done elsewhere (when players enter range) (tick?)
		for _, p := range players {
			if p.UUID() != player.UUID() {
				buf.Write(clientbound.SpawnPlayer{
					EntityID:   pk.VarInt(p.EntityID()),
					PlayerUUID: pk.UUID(p.UUID()),
					X:          pk.Double(p.Position().X),
					Y:          pk.Double(p.Position().Y),
					Z:          pk.Double(p.Position().Z),
					Yaw:        pk.Angle(p.Rotation().Yaw / 360 * 256),
					Pitch:      pk.Angle(p.Rotation().Pitch / 360 * 256),
				}.CreatePacket().Encode())
			}
		}

		out = buf.Bytes()

		event := events.PlayerJoinData{
			Player:  &player,
			Message: fmt.Sprintf("%v has joined the game", player.Name()),
		}
		events.PlayerJoinEvent.Trigger(&event)

		// send out player info to players online
		playerInfoPacket := clientbound.PlayerInfo{
			Action:     0,
			NumPlayers: 1,
			Players: []pk.Encodable{
				clientbound.PlayerInfoAddPlayer{
					UUID:           pk.UUID(player.UUID()),
					Name:           pk.String(player.Name()),
					NumProperties:  0,
					Properties:     nil,
					Gamemode:       0,
					Ping:           0,
					HasDisplayName: false,
					DisplayName:    "",
				},
			},
		}.CreatePacket().Encode()
		// TODO: spawn player should be occurring when players enter range (not join game), do logic elsewhere (tick?)
		spawnPlayerPacket := clientbound.SpawnPlayer{
			EntityID:   pk.VarInt(player.EntityID()),
			PlayerUUID: pk.UUID(player.UUID()),
			X:          pk.Double(player.Position().X),
			Y:          pk.Double(player.Position().Y),
			Z:          pk.Double(player.Position().Z),
			Yaw:        pk.Angle(player.Rotation().Yaw / 360 * 256),
			Pitch:      pk.Angle(player.Rotation().Pitch / 360 * 256),
		}.CreatePacket().Encode()

		for _, p := range s.Players() {
			conn := s.ConnFromUUID(p.UUID())
			if conn != c { // Don't send spawn player packets to self...
				_ = conn.AsyncWrite(append(playerInfoPacket, spawnPlayerPacket...))
			}
		}

		s.Broadcast(event.Message)

	} else {
		// TODO: send kick message
		return nil, errors.New("login not allowed not yet implemented")
	}

	return
}
