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

func ChatMessage(c gnet.Conn, pkt *pk.Packet, s api.Server) (out []byte, err error) {
	m := serverbound.ChatMessage{}
	if err = m.FromPacket(pkt); err != nil {
		return
	}
	player := s.PlayerFromConn(c)
	logger.Printf("Received chat message `%v` from %v", m.Message, player.GetName())

	// create message event
	msg := chat.NewMessage(fmt.Sprintf("%s: %s", player.GetName(), m.Message))
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
		Sender:   pk.UUID((*event.Player).GetUUID()),
	}.CreatePacket().Encode()
	for _, p := range event.Recipients {
		conn := s.ConnFromUUID(p.GetUUID())
		if conn != c {	// don't send to self: React method will take out bytes and send them
			_ = conn.AsyncWrite(out)
		}
	}
	return
}
