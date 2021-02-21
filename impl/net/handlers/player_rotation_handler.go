package handlers

import (
	"github.com/panjf2000/gnet"
	"gogs/api"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/serverbound"
)

func PlayerRotation(c gnet.Conn, pkt *pk.Packet, s api.Server) error {
	player := s.PlayerFromConn(c)
	logger.Printf("Received player rotation for %s", player.Name)

	rotationPacket := serverbound.PlayerRotation{}
	if err := rotationPacket.FromPacket(pkt); err != nil {
		return err
	}

	// send entity rotation packet
	outPacket := clientbound.EntityRotation{
		EntityID: pk.VarInt(player.EntityID),
		Yaw:      pk.Angle(rotationPacket.Yaw / 360 * 256),
		Pitch:    pk.Angle(rotationPacket.Pitch / 360 * 256),
		OnGround: rotationPacket.OnGround,
	}.CreatePacket().Encode()

	// also send head rotation packet
	outPacket = append(outPacket, clientbound.EntityHeadLook{
		EntityID: pk.VarInt(player.EntityID),
		HeadYaw:  pk.Angle(rotationPacket.Yaw / 360 * 256),
	}.CreatePacket().Encode()...)

	for _, p := range s.Players() {
		conn := s.ConnFromUUID(p.UUID)
		if conn != c {
			_ = conn.AsyncWrite(outPacket)
		}
	}
	return nil
}
