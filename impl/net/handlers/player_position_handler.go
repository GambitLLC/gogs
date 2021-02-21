package handlers

import (
	"github.com/panjf2000/gnet"
	"gogs/api"
	"gogs/api/game"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/serverbound"
)

func PlayerPosition(c gnet.Conn, pkt *pk.Packet, s api.Server) error {
	player := s.PlayerFromConn(c)
	logger.Printf("Received player position for %v", player.Name)
	pos := serverbound.PlayerPosition{}
	if err := pos.FromPacket(pkt); err != nil {
		return err
	}

	// TODO: why doesn't this work
	// send new player position to everyone else
	playerPositionPacket := clientbound.EntityPosition{
		EntityID: pk.VarInt(player.EntityID),
		DeltaX:   pk.Short(float64(pos.X) - player.Position.X),
		DeltaY:   pk.Short(float64(pos.Y) - player.Position.Y),
		DeltaZ:   pk.Short(float64(pos.Z) - player.Position.Z),
		OnGround: pos.OnGround,
	}.CreatePacket().Encode()
	for _, p := range s.Players() {
		conn := s.ConnFromUUID(p.UUID)
		if conn != c {
			logger.Printf("Sending position for %v to %v", player.Name, p.Name)
			_ = conn.AsyncWrite(playerPositionPacket)
		}
	}

	// update player position
	player.Position = game.Position{
		X: float64(pos.X),
		Y: float64(pos.Y),
		Z: float64(pos.Z),
	}

	return nil
}
