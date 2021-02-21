package listeners

import (
	"errors"
	"github.com/panjf2000/gnet"
	"gogs/api"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/packetids"
	"gogs/impl/net/packet/serverbound"
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

		return nil, nil

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
