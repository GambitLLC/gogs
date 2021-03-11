package server

import (
	"github.com/panjf2000/gnet"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/serverbound"
)

func (s *Server) handleClientSettings(conn gnet.Conn, pkt pk.Packet) (out []byte, err error) {
	logger.Printf("Received client settings")
	in := serverbound.ClientSettings{}
	if err = in.FromPacket(pkt); err != nil {
		return
	}

	player := s.playerFromConn(conn)
	player.ChatMode = uint8(in.ChatMode)
	player.ViewDistance = byte(in.ViewDistance)

	return
}
