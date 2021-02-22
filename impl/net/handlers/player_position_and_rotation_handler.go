package handlers

import (
	"github.com/panjf2000/gnet"
	"gogs/api"
	"gogs/api/data"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/serverbound"
)

func PlayerPositionAndRotation(c gnet.Conn, pkt *pk.Packet, s api.Server) error {
	player := s.PlayerFromConn(c)
	logger.Printf("Received player pos and rotation from %v", player.GetName())
	in := serverbound.PlayerPositionAndRotation{}
	if err := in.FromPacket(pkt); err != nil {
		return err
	}

	player.SetPosition(data.Position{
		X:        float64(in.X),
		Y:        float64(in.Y),
		Z:        float64(in.Z),
		OnGround: bool(in.OnGround),
	})

	player.SetRotation(data.Rotation{
		Yaw:      float32(in.Yaw),
		Pitch:    float32(in.Pitch),
		OnGround: bool(in.OnGround),
	})

	//outPacket := clientbound.EntityPositionAndRotation{
	//	EntityID: pk.VarInt(player.GetEntityID()),
	//	DeltaX:   pk.Short(float64(data.X*32)-player.GetPosition().X*32) * 128,
	//	DeltaY:   pk.Short(float64(data.Y*32)-player.GetPosition().Y*32) * 128,
	//	DeltaZ:   pk.Short(float64(data.Z*32)-player.GetPosition().Z*32) * 128,
	//	Yaw:      pk.Angle(data.Yaw / 360 * 256),
	//	Pitch:    pk.Angle(data.Pitch / 360 * 256),
	//	OnGround: data.OnGround,
	//}.CreatePacket().Encode()
	//
	//// also send head rotation packet
	//outPacket = append(outPacket, clientbound.EntityHeadLook{
	//	EntityID: pk.VarInt(player.GetEntityID()),
	//	HeadYaw:  pk.Angle(data.Yaw / 360 * 256),
	//}.CreatePacket().Encode()...)
	//
	//for _, p := range s.Players() {
	//	conn := s.ConnFromUUID(p.GetUUID())
	//	if conn != c {
	//		_ = conn.AsyncWrite(outPacket)
	//	}
	//}

	return nil
}
