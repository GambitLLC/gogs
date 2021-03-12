package server

import (
	"bytes"
	"fmt"
	"github.com/panjf2000/gnet"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
)

func (s *Server) handleClientStatus(conn gnet.Conn, pkt pk.Packet) ([]byte, error) {
	var actionID pk.VarInt
	if err := pkt.Unmarshal(&actionID); err != nil {
		return nil, err
	}

	switch actionID {
	case 0: // Perform respawn
		player := s.playerFromConn(conn)
		player.Health = 20
		buf := bytes.Buffer{}
		// send respawn packet
		buf.Write(clientbound.Respawn{
			Dimension:        pk.NBT{V: clientbound.MinecraftOverworld},
			WorldName:        "world",
			HashedSeed:       0,
			Gamemode:         0,
			PreviousGamemode: 0,
			IsDebug:          false,
			IsFlat:           true,
			CopyMetadata:     false,
		}.CreatePacket().Encode())
		// TODO: send inventory
		return buf.Bytes(), nil
	case 1: // Request stats
	default:
		return nil, fmt.Errorf("client status got invalid action id %d", actionID)
	}

	return nil, nil
}
