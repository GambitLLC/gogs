package handlers

import (
	"github.com/panjf2000/gnet"
	"gogs/api"
	"gogs/api/data"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/serverbound"
)

func PlayerPosition(c gnet.Conn, pkt *pk.Packet, s api.Server) error {
	player := s.PlayerFromConn(c)
	logger.Printf("Received player position for %v", player.GetName())
	pos := serverbound.PlayerPosition{}
	if err := pos.FromPacket(pkt); err != nil {
		return err
	}

	// update player position
	player.SetPosition(data.Position{
		X:        float64(pos.X),
		Y:        float64(pos.Y),
		Z:        float64(pos.Z),
		OnGround: bool(pos.OnGround),
	})

	return nil
}
