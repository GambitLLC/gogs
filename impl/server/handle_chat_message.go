package server

import (
	"fmt"
	"gogs/api/data/chat"
	"gogs/impl/logger"
	"gogs/impl/net"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/serverbound"
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
		s.shuttingDown = true
	}

	msg := chat.NewMessage(fmt.Sprintf("%s: %s", player.Name, m.Message))
	tmp := clientbound.ChatMessage{
		JSONData: pk.Chat(msg.AsJSON()),
		Position: 0,
		Sender:   pk.UUID(player.UUID),
	}.CreatePacket()

	s.playerMap.Lock.RLock()
	defer s.playerMap.Lock.RUnlock()
	for _, p := range s.playerMap.uuidToPlayer {
		_ = p.Connection.WritePacket(tmp)
	}

	/*
		players := make([]api.Player, len(s.playerMap.uuidToPlayer))
		for _, p := range s.playerMap.uuidToPlayer {
			players = append(players, api.Player(p))
		}

		// create message event
		msg := chat.NewMessage(fmt.Sprintf("%s: %s", player.Name, m.Message))
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
