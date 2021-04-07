package server

import (
	"gogs/chat"
	"gogs/logger"
	"gogs/net"
	pk "gogs/net/packet"
	"gogs/net/packet/clientbound"
	"gogs/net/packet/serverbound"
)

func (s *Server) handleChatMessage(conn net.Conn, pkt pk.Packet) (err error) {
	m := serverbound.ChatMessage{}
	if err = m.FromPacket(pkt); err != nil {
		return
	}
	player := s.playerFromConn(conn)
	logger.Printf("Received chat message `%v` from %v", m.Message, player.Name)

	// TODO: MOVE THIS INTO COMMAND HANDLER
	if m.Message == "/stop" {
		s.stop()
		return
	}

	msg := chat.NewTranslationComponent(
		"chat.type.text", // "<%s> %s"
		chat.NewStringComponent(player.Name),
		chat.NewStringComponent(string(m.Message)),
	)

	tmp := clientbound.ChatMessage{
		JSONData: pk.Chat(msg.AsJSON()),
		Position: chat.Chat,
		Sender:   pk.UUID(player.UUID),
	}.CreatePacket()

	s.playerMap.Lock.RLock()
	defer s.playerMap.Lock.RUnlock()
	for _, p := range s.playerMap.connToPlayer {
		_ = p.Connection.WritePacket(tmp)
	}

	/*
		players := make([]api.Player, len(s.playerMap.uuidToPlayer))
		for _, p := range s.playerMap.uuidToPlayer {
			players = append(players, api.Player(p))
		}

		// create message event
		msg := chat.NewStringComponent(fmt.Sprintf("%s: %s", player.Name, m.Message))
		event := events.PlayerChatData{
			Player:     player,
			Recipients: players,
			Message:    msg,
		}
		events.PlayerChatEvent.Trigger(&event)

		// send chat message to the recipients
		out = clientbound.ChatMessage{
			JSONData: pk.Chat(event.Message.AsJSON()),
			Position: 0,
			Sender:   pk.UUID(event.Player.UUID()),
		}.CreatePacket().Encode()
		for _, p := range event.Recipients {
			c := s.connFromUUID(p.UUID())
			_ = c.AsyncWrite(out)
		}
	*/
	return
}
