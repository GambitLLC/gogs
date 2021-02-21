package listeners

import (
	"bytes"
	"errors"
	"github.com/google/uuid"
	"github.com/panjf2000/gnet"
	"gogs/api"
	"gogs/api/events"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/packetids"
)

type LoginState int8

const (
	start LoginState = iota
	encrypt
)

type LoginPacketListener struct {
	S               api.Server
	protocolVersion int32
	encrypt         bool
	state           LoginState
}

func (listener LoginPacketListener) HandlePacket(c gnet.Conn, p *pk.Packet) ([]byte, error) {
	switch listener.state {
	case start:
		return listener.handleLoginStart(c, p)
	case encrypt:
		return nil, errors.New("login encryption is not yet implemented")
	default:
		logger.Printf("LoginPacketListener is in an unknown state: %d", listener.state)
		return nil, c.Close()
	}
}

func (listener *LoginPacketListener) handleLoginStart(c gnet.Conn, p *pk.Packet) ([]byte, error) {
	if p.ID != 0 {
		return nil, errors.New("login start expects Packet ID 0")
	}

	var name pk.String

	if err := p.Unmarshal(&name); err != nil {
		return nil, err
	}

	logger.Printf("Received login start packet from player %v", name)

	if len(name) > 16 {
		// send disconnect
		return pk.Marshal(packetids.LoginDisconnect, pk.Chat("username too long")).Encode(), errors.New("username too long")
	}

	// TODO: send encryption request
	if listener.encrypt {
		/*
			out = pk.Marshal(
				0x01,
				pk.String(""),    // Server ID
				pk.VarInt(1),    // public key length
				pk.ByteArray([]byte("s")), // public key in bytes
				pk.VarInt(1),    // verify token length
				pk.ByteArray([]byte("s")), // verify token in bytes
			).Encode()
		*/
		return nil, errors.New("encryption (online mode) is not implemented")
	} else {
		c.SetContext(PlayPacketListener{
			S:               listener.S,
			protocolVersion: listener.protocolVersion,
		})

		player := listener.S.CreatePlayer(string(name), uuid.UUID(pk.NameToUUID(string(name))), c)

		buf := bytes.Buffer{}

		//trigger login event
		events.PlayerLoginEvent.Trigger(&events.PlayerLoginData{
			Player: player,
			Conn:   c,
		})

		// TODO: move this triggering into after (if) login event is successful
		// trigger player join event
		events.PlayerJoinEvent.Trigger(&events.PlayerJoinData{
			Player:  player,
			Message: "",
		})

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

		return buf.Bytes(), nil
	}
}
