package server

import (
	"github.com/panjf2000/gnet"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/serverbound"
)

func (s *Server) handlePlayerPositionAndRotation(conn gnet.Conn, pkt pk.Packet) (out []byte, err error) {
	player := s.playerFromConn(conn)
	logger.Printf("Received player pos and rotation from %v", player.Name)
	in := serverbound.PlayerPositionAndRotation{}
	if err = in.FromPacket(pkt); err != nil {
		return
	}

	outPacket := clientbound.EntityPositionAndRotation{
		EntityID: pk.VarInt(player.ID()),
		DeltaX:   pk.Short((float64(in.X*32) - player.X*32) * 128),
		DeltaY:   pk.Short((float64(in.Y*32) - player.Y*32) * 128),
		DeltaZ:   pk.Short((float64(in.Z*32) - player.Z*32) * 128),
		Yaw:      pk.Angle(in.Yaw / 360 * 256),
		Pitch:    pk.Angle(in.Pitch / 360 * 256),
		OnGround: in.OnGround,
	}.CreatePacket()
	s.broadcastPacket(outPacket, conn)

	// also send head rotation packet
	outPacket = clientbound.EntityHeadLook{
		EntityID: pk.VarInt(player.ID()),
		HeadYaw:  pk.Angle(in.Yaw / 360 * 256),
	}.CreatePacket()
	s.broadcastPacket(outPacket, conn)

	player.X = float64(in.X)
	player.Y = float64(in.Y)
	player.Z = float64(in.Z)

	player.Yaw = float32(in.Yaw)
	player.Pitch = float32(in.Pitch)

	return nil, nil
}
