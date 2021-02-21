package handlers

import (
	"gogs/api"
	"gogs/api/events"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
)

func PlayerChatHandler(s api.Server) func(data *events.PlayerChatData) {
	return func(data *events.PlayerChatData) {
		msg := clientbound.ChatMessage{
			JSONData: pk.Chat(data.AsJSON()),
			Position: 0,
			Sender:   pk.UUID(data.Player.UUID),
		}.CreatePacket().Encode()
		for _, p := range data.Recipients {
			c := s.ConnFromUUID(p.UUID)
			c.AsyncWrite(msg)
		}
	}
}
