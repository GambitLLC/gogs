package listeners

import (
	"errors"
	"fmt"
	"github.com/panjf2000/gnet"
	"gogs/api"
	"gogs/api/data/chat"
	"gogs/api/events"
	"gogs/impl/logger"
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
	case packetids.TeleportConfirm:
		// TODO: Handle this
		logger.Printf("Received teleport confirm")
	case packetids.ChatMessageServerbound:
		s := serverbound.ChatMessage{}
		if err := s.FromPacket(p); err != nil {
			return nil, err
		}
		player := listener.S.PlayerFromConn(c)
		logger.Printf("Received chat message `%v` from %v", s.Message, player.Name)
		msg := chat.NewMessage(fmt.Sprintf("%s: %s", player.Name, s.Message))
		events.PlayerChatEvent.Trigger(&events.PlayerChatData{
			Player:     player,
			Recipients: listener.S.Players(),
			Message:    msg,
		})

	case packetids.ClientSettings:
		logger.Printf("Received client settings")
		// TODO: actually handle client settings
		s := serverbound.ClientSettings{}
		if err := s.FromPacket(p); err != nil {
			return nil, err
		}
	case packetids.PlayerPosition:
		// TODO: Handle all player pos & rotation packets
		logger.Printf("Received player position")
	case packetids.PlayerPositionAndRotationServerbound:
		logger.Printf("Received player pos and rotation")
		s := serverbound.PlayerPositionAndRotation{}
		if err := s.FromPacket(p); err != nil {
			return nil, err
		}
	case packetids.PlayerRotation:
		logger.Printf("Received player rotation")
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

	return nil, nil
}
