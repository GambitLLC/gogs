package clientbound

import (
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/packetids"
)

type SpawnLivingEntity struct {
	EntityID   pk.VarInt
	EntityUUID pk.UUID
	Type       pk.VarInt
	X          pk.Double
	Y          pk.Double
	Z          pk.Double
	Yaw        pk.Angle
	Pitch      pk.Angle
	HeadPitch  pk.Angle
	VelocityX  pk.Short
	VelocityY  pk.Short
	VelocityZ  pk.Short
}

func (s SpawnLivingEntity) CreatePacket() pk.Packet {
	return pk.Marshal(
		packetids.SpawnLivingEntity,
		s.EntityID,
		s.EntityUUID,
		s.Type,
		s.X,
		s.Y,
		s.Z,
		s.Yaw,
		s.Pitch,
		s.HeadPitch,
		s.VelocityX,
		s.VelocityY,
		s.VelocityZ,
	)
}
