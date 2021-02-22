package handlers

import (
	"github.com/panjf2000/gnet"
	"gogs/api"
	"gogs/api/data"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/serverbound"
)

func PlayerRotation(c gnet.Conn, pkt *pk.Packet, s api.Server) error {
	player := s.PlayerFromConn(c)
	logger.Printf("Received player rotation for %s", player.GetName())

	rotationPacket := serverbound.PlayerRotation{}
	if err := rotationPacket.FromPacket(pkt); err != nil {
		return err
	}

	player.SetRotation(data.Rotation{
		Yaw:      float32(rotationPacket.Yaw),
		Pitch:    float32(rotationPacket.Pitch),
		OnGround: bool(rotationPacket.OnGround),
	})

	return nil
}
