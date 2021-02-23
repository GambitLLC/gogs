package server

import (
	"github.com/panjf2000/gnet"
	"gogs/api/data"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/serverbound"
)

func (s *Server) handlePlayerPosition(conn gnet.Conn, pkt pk.Packet) (out []byte, err error) {
	player := s.PlayerFromConn(conn)
	logger.Printf("Received player position for %v", player.Name())
	in := serverbound.PlayerPosition{}
	if err := in.FromPacket(pkt); err != nil {
		return nil, err
	}

	outPacket := clientbound.EntityPosition{
		EntityID: pk.VarInt(player.EntityID()),
		DeltaX:   pk.Short((float64(in.X*32) - player.Position().X*32) * 128),
		DeltaY:   pk.Short((float64(in.Y*32) - player.Position().Y*32) * 128),
		DeltaZ:   pk.Short((float64(in.Z*32) - player.Position().Z*32) * 128),
		OnGround: in.OnGround,
	}.CreatePacket()

	s.broadcastPacket(outPacket, conn)

	// update player position
	*player.Position() = data.Position{
		X:        float64(in.X),
		Y:        float64(in.Y),
		Z:        float64(in.Z),
		OnGround: bool(in.OnGround),
	}

	return nil, nil
}
