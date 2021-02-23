package server

import (
	"fmt"
	"github.com/panjf2000/gnet"
	"gogs/api/data/chat"
	"gogs/api/events"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/serverbound"
)

func (s *Server) handleChatMessage(conn gnet.Conn, pkt pk.Packet) (out []byte, err error) {
	m := serverbound.ChatMessage{}
	if err = m.FromPacket(pkt); err != nil {
		return
	}
	player := s.PlayerFromConn(conn)
	logger.Printf("Received chat message `%v` from %v", m.Message, player.Name())

	// create message event
	msg := chat.NewMessage(fmt.Sprintf("%s: %s", player.Name(), m.Message))
	event := events.PlayerChatData{
		Player:     &player,
		Recipients: s.Players(),
		Message:    msg,
	}
	events.PlayerChatEvent.Trigger(&event)

	// send chat message to the recipients
	out = clientbound.ChatMessage{
		JSONData: pk.Chat(event.Message.AsJSON()),
		Position: 0,
		Sender:   pk.UUID((*event.Player).UUID()),
	}.CreatePacket().Encode()
	for _, p := range event.Recipients {
		c := s.ConnFromUUID(p.UUID())
		_ = c.AsyncWrite(out)
	}
	return nil, nil
}
