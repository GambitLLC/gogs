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

func LoginStart(c gnet.Conn, p *pk.Packet, s api.Server) error {
	if p.ID != 0 {
		return errors.New("login start expects Packet ID 0")
	}

	var name pk.String
	if err := p.Unmarshal(&name); err != nil {
		return err
	}

	logger.Printf("Received login start packet from player %v", name)

	//trigger login event
	event := events.PlayerLoginData{
		Name: string(name),
		Conn: c,
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
			Gamemode:     1,
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

		buf.Write((&clientbound.PlayerPositionAndLook{}).FromPlayer(*player).CreatePacket().Encode())

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

		buf.Write((&clientbound.PlayerPositionAndLook{}).FromPlayer(*player).CreatePacket().Encode())

		if err := c.AsyncWrite(buf.Bytes()); err != nil {
			return err
		}

		event := events.PlayerJoinData{
			Player:  player,
			Message: fmt.Sprintf("%v has joined the game", player.Name),
		}
		events.PlayerJoinEvent.Trigger(&event)

		s.Broadcast(event.Message)

	} else {
		// TODO: send kick message
		return errors.New("login not allowed not yet implemented")
	}

	return nil
}
