package clientbound

import (
	pk "gogs/net/packet"
	"gogs/net/packet/packetids"
)

type SpawnEntity struct {
	EntityID   pk.VarInt
	ObjectUUID pk.UUID
	Type       pk.VarInt
	X          pk.Double
	Y          pk.Double
	Z          pk.Double
	Pitch      pk.Angle
	Yaw        pk.Angle
	Data       pk.Int
	VelocityX  pk.Short
	VelocityY  pk.Short
	VelocityZ  pk.Short
}

func (s SpawnEntity) CreatePacket() pk.Packet {
	return pk.Marshal(
		packetids.SpawnEntity,
		s.EntityID,
		s.ObjectUUID,
		s.Type,
		s.X,
		s.Y,
		s.Z,
		s.Pitch,
		s.Yaw,
		s.Data,
		s.VelocityX,
		s.VelocityY,
		s.VelocityZ,
	)
}
