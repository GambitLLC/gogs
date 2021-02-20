package listeners

import (
	"bytes"
	"errors"
	"github.com/panjf2000/gnet"
	"gogs/api"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/packetids"
	"gogs/impl/net/packet/serverbound"
	"log"
)

type PlayPacketListener struct {
	S               api.Server
	protocolVersion int32
}

func (listener PlayPacketListener) HandlePacket(c gnet.Conn, p *pk.Packet) ([]byte, error) {
	switch p.ID {
	case packetids.ClientSettings:
		s := serverbound.ClientSettings{}
		if err := s.FromPacket(p); err != nil {
			return nil, err
		}
		// https://wiki.vg/Protocol_FAQ#What.27s_the_normal_login_sequence_for_a_client.3F
		buf := bytes.Buffer{}
		buf.Write(clientbound.HeldItemChange{}.CreatePacket().Encode())

		buf.Write(clientbound.DeclareRecipes{
			NumRecipes: 0,
			Recipes:    nil,
		}.CreatePacket().Encode())

		player := listener.S.PlayerFromConn(c)

		buf.Write((&clientbound.PlayerPositionAndLook{}).FromPlayer(*player).CreatePacket().Encode())

		buf.Write(clientbound.PlayerInfo{
			Action:     0,
			NumPlayers: 1,
			Players: []pk.Encodable{
				clientbound.PlayerInfoAddPlayer{
					UUID:           pk.UUID(player.UUID),
					Name:           pk.String(player.Name),
					NumProperties:  pk.VarInt(0),
					Properties:     nil,
					Gamemode:       pk.VarInt(0),
					Ping:           pk.VarInt(0),
					HasDisplayName: false,
					DisplayName:    "",
				},
			},
		}.CreatePacket().Encode())

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
				log.Print("creating chunk data")
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
				log.Print(chunk)
				//log.Print("Sending chunk data")
				buf.Write(chunk)
			}
		}

		buf.Write(clientbound.SpawnPosition{Location: pk.Position{
			X: 0,
			Y: 2,
			Z: 0,
		}}.CreatePacket().Encode())

		buf.Write((&clientbound.PlayerPositionAndLook{}).FromPlayer(*player).CreatePacket().Encode())

		logger.Printf("Writing back")
		return buf.Bytes(), nil

	case packetids.PlayerPositionAndLook:
		s := serverbound.PlayerPositionAndLook{}
		if err := s.FromPacket(p); err != nil {
			return nil, err
		}
	case 0x10:
		//TODO: kick client for incorrect / untimely Keep-Alive response
		s := serverbound.KeepAlive{}
		if err := s.FromPacket(p); err != nil {
			return nil, err
		}

	default:
		return nil, errors.New("not yet implemented")
	}

	return nil, nil
}
