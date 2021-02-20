package listeners

import (
	"errors"
	"github.com/panjf2000/gnet"
	pk "gogs/net/packet"
	"gogs/net/ptypes"
	"log"
)

type LoginState int8

const (
	start LoginState = iota
	encrypt
)

type loginPacketListener struct {
	protocolVersion int32
	encrypt bool
	state LoginState
}

func LoginPacketListener(protoVersion int32, encrypt bool) loginPacketListener {
	return loginPacketListener{
		protocolVersion: protoVersion,
		encrypt: 		 encrypt,
		state:           start,
	}
}

func (listener loginPacketListener) HandlePacket(c gnet.Conn, p *pk.Packet) error {
	switch listener.state {
	case start:
		return listener.handleLoginStart(c, p)
	case encrypt:
		return errors.New("not yet implemented")
	default:
		log.Panicf("Unhandled state in loginPacketListener: %v", listener.state)
	}
	return nil
}

func (listener *loginPacketListener) handleLoginStart(c gnet.Conn, p *pk.Packet) error {
	if p.ID != 0 {
		return errors.New("login start expects Packet ID 0")
	}

	var name pk.String

	if err := p.Unmarshal(&name); err != nil {
		return err
	}

	log.Printf("received login from player %v", name)

	if len(name) > 16 {
		// TODO: define packetid consts and use them
		// send disconnect
		c.SendTo(pk.Marshal(0x00, pk.Chat("username too long")).Encode())
		return errors.New("username too long")
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
		return errors.New("encryption (online mode) is not implemented")
	} else {
		c.SetContext(PlayPacketListener(listener.protocolVersion))
		// send login success (offline mode for now)
		c.SendTo(pk.Marshal(
			0x02,
			pk.UUID(pk.NameToUUID(string(name))), // UUID
			pk.String(name),                      // Username
		).Encode())

		// also send out join game
		c.SendTo(ptypes.JoinGame{
			PlayerEntity: 12193,
			Hardcore:     false,
			Gamemode:     0,
			PrevGamemode: 0,
			WorldCount:   1,
			WorldNames:   []pk.Identifier{"world"},
			DimensionCodec: pk.NBT{
				V: ptypes.DimensionCodec{
					DimensionTypes: ptypes.DimensionTypeRegistry{
						Type: "minecraft:dimension_type",
						Value: []ptypes.DimensionTypeRegistryEntry{
							{"minecraft:overworld",
								0,
								ptypes.MinecraftOverworld,
							},
						},
					},
					BiomeRegistry: ptypes.BiomeRegistry{
						Type:  "minecraft:worldgen/biome",
						Value: []ptypes.BiomeRegistryEntry{
							{
								Name: "minecraft:plains",
								ID:   1,
								Element: ptypes.BiomeProperties{
									Precipitation: "none",
									Depth:         0.125,
									Temperature:   0.8,
									Scale:         0.05,
									Downfall:      0.4,
									Category:      "plains",
									Effects: ptypes.BiomeEffects{
										SkyColor:      7907327,
										WaterFogColor: 329011,
										FogColor:      12638463,
										WaterColor:    4159204,
									},
								},
							},
						},
					},
				}},
			Dimension:    pk.NBT{V: ptypes.MinecraftOverworld},
			WorldName:    "world",
			HashedSeed:   0,
			MaxPlayers:   20,
			ViewDistance: 10,
			RDI:          false,
			ERS:          false,
			IsDebug:      false,
			IsFlat:       false,
		}.CreatePacket().Encode())
	}

	return nil
}