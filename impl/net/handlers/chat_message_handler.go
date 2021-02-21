package handlers

import (
	"fmt"
	"github.com/panjf2000/gnet"
	"gogs/api"
	"gogs/api/data/chat"
	"gogs/api/events"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/serverbound"
)

func ChatMessage(c gnet.Conn, p *pk.Packet, s api.Server) error {
	m := serverbound.ChatMessage{}
	if err := m.FromPacket(p); err != nil {
		return err
	}
	player := s.PlayerFromConn(c)
	logger.Printf("Received chat message `%v` from %v", m.Message, player.Name)

	// create message event
	msg := chat.NewMessage(fmt.Sprintf("%s: %s", player.Name, m.Message))
	event := events.PlayerChatData{
		Player:     player,
		Recipients: s.Players(),
		Message:    msg,
	}
	events.PlayerChatEvent.Trigger(&event)

	// send chat message to the recipients
	packet := clientbound.ChatMessage{
		JSONData: pk.Chat(event.Message.AsJSON()),
		Position: 0,
		Sender:   pk.UUID(event.Player.UUID),
	}.CreatePacket().Encode()
	for _, p := range event.Recipients {
		c := s.ConnFromUUID(p.UUID)
		if err := c.AsyncWrite(packet); err != nil {
			return err
		}
	}
	return nil
}
