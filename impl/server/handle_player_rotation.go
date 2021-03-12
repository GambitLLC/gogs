package server

import (
	"github.com/panjf2000/gnet"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/serverbound"
)

func (s *Server) handlePlayerRotation(conn gnet.Conn, pkt pk.Packet) (out []byte, err error) {
	player := s.playerFromConn(conn)
	logger.Printf("Received player rotation for %s", player.Name)

	in := serverbound.PlayerRotation{}
	if err = in.FromPacket(pkt); err != nil {
		return
	}

	outPacket := clientbound.EntityRotation{
		EntityID: pk.VarInt(player.ID()),
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

	player.Yaw = float32(in.Yaw)
	player.Pitch = float32(in.Pitch)

	return
}
