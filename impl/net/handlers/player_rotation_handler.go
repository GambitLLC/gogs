package handlers

import (
	"github.com/panjf2000/gnet"
	"gogs/api"
	"gogs/api/data"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/serverbound"
)

func PlayerRotation(c gnet.Conn, pkt *pk.Packet, s api.Server) ([]byte, error) {
	player := s.PlayerFromConn(c)
	logger.Printf("Received player rotation for %s", player.Name())

	in := serverbound.PlayerRotation{}
	if err := in.FromPacket(pkt); err != nil {
		return nil, err
	}

	outPacket := clientbound.EntityRotation{
		EntityID: pk.VarInt(player.EntityID()),
		Yaw:      pk.Angle(in.Yaw / 360 * 256),
		Pitch:    pk.Angle(in.Pitch / 360 * 256),
		OnGround: in.OnGround,
	}.CreatePacket().Encode()
	// also send head rotation packet
	outPacket = append(outPacket, clientbound.EntityHeadLook{
		EntityID: pk.VarInt(player.EntityID()),
		HeadYaw:  pk.Angle(in.Yaw / 360 * 256),
	}.CreatePacket().Encode()...)

	for _, player := range s.Players() {
		conn := s.ConnFromUUID(player.UUID())
		if conn != c {
			_ = conn.AsyncWrite(outPacket)
		}
	}

	*player.Rotation() = data.Rotation{
		Yaw:      float32(in.Yaw),
		Pitch:    float32(in.Pitch),
		OnGround: bool(in.OnGround),
	}

	return nil, nil
}
