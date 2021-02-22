package listeners

import (
	"errors"
	"github.com/panjf2000/gnet"
	"gogs/api"
	"gogs/impl/logger"
	"gogs/impl/net/handlers"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/packetids"
	"gogs/impl/net/packet/serverbound"
)

type PlayPacketListener struct {
	S               api.Server
	protocolVersion int32
}

func (listener PlayPacketListener) HandlePacket(c gnet.Conn, p *pk.Packet) (out []byte, err error) {
	switch p.ID {
	case packetids.TeleportConfirm:
		// TODO: Handle this
		logger.Printf("Received teleport confirm")
	case packetids.ChatMessageServerbound:
		if out, err = handlers.ChatMessage(c, p, listener.S); err != nil {
			return nil, err
		}
	case packetids.ClientSettings:
		logger.Printf("Received client settings")
		// TODO: actually handle client settings
		s := serverbound.ClientSettings{}
		if err := s.FromPacket(p); err != nil {
			return nil, err
		}
	case packetids.PlayerPosition:
		// TODO: Handle all player pos & rotation packets
		return handlers.PlayerPosition(c, p, listener.S)
	case packetids.PlayerPositionAndRotationServerbound:
		return handlers.PlayerPositionAndRotation(c, p, listener.S)
	case packetids.PlayerRotation:
		return handlers.PlayerRotation(c, p, listener.S)
	case packetids.Animation:
		return handlers.Animation(c, p, listener.S)
	case packetids.EntityAction:
		return handlers.EntityAction(c, p, listener.S)
	case packetids.KeepAliveServerbound:
		logger.Printf("Received keep alive")
		//TODO: kick client for incorrect / untimely Keep-Alive response
		s := serverbound.KeepAlive{}
		if err := s.FromPacket(p); err != nil {
			return nil, err
		}

	default:
		return nil, errors.New("not yet implemented")
	}

	return
}
