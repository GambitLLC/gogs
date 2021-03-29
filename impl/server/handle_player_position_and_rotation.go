package server

import (
	"gogs/impl/logger"
	"gogs/impl/net"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/serverbound"
)

func (s *Server) handlePlayerPositionAndRotation(conn net.Conn, pkt pk.Packet) (err error) {
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

	// if chunk border was crossed, update view pos and send new chunks
	if int(player.X)>>4 != int(in.X)>>4 || int(player.Z)>>4 != int(in.Z)>>4 {
		err = s.updateViewPosition(player)
	}

	player.X = float64(in.X)
	player.Y = float64(in.Y)
	player.Z = float64(in.Z)

	player.Yaw = float32(in.Yaw)
	player.Pitch = float32(in.Pitch)

	return
}
