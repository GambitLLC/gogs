package listeners

import (
	"bytes"
	"errors"
	"github.com/google/uuid"
	"github.com/panjf2000/gnet"
	"gogs/api"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"log"
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
		return nil, errors.New("not yet implemented")
	default:
		log.Panicf("Unhandled state in LoginPacketListener: %v", listener.state)
	}
	return nil, nil
}

func (listener *LoginPacketListener) handleLoginStart(c gnet.Conn, p *pk.Packet) ([]byte, error) {
	if p.ID != 0 {
		return nil, errors.New("login start expects Packet ID 0")
	}

	var name pk.String

	if err := p.Unmarshal(&name); err != nil {
		return nil, err
	}

	logger.Printf("received login from player %v", name)

	if len(name) > 16 {
		// TODO: define packetid consts and use them
		// send disconnect
		return pk.Marshal(0x00, pk.Chat("username too long")).Encode(), errors.New("username too long")
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

		// send login success
		buf := bytes.Buffer{}
		buf.Write(pk.Marshal(
			0x02,
			pk.UUID(player.UUID),
			pk.String(player.Name),
		).Encode())

		// trigger login event
		//events.PlayerLoginEvent.Trigger(&events.PlayerLoginData{
		//	Player: player,
		//	Conn:   c,
		//})
		//
		//events.PlayerJoinEvent.Trigger(&events.PlayerJoinData{
		//	Player:  player,
		//	Message: "",
		//})

		buf.Write(clientbound.JoinGame{
			PlayerEntity: 12193,
			Hardcore:     false,
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
										SkyColor:      0x00FF00,
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

		return buf.Bytes(), nil
	}

	return nil, nil
}
