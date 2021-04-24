package clientbound

import (
	pk "github.com/GambitLLC/gogs/net/packet"
	"github.com/GambitLLC/gogs/net/packet/packetids"
)

type EntityPositionAndRotation struct {
	EntityID pk.VarInt
	DeltaX   pk.Short
	DeltaY   pk.Short
	DeltaZ   pk.Short
	Yaw      pk.Angle
	Pitch    pk.Angle
	OnGround pk.Boolean
}

func (s EntityPositionAndRotation) CreatePacket() pk.Packet {
	return pk.Marshal(
		packetids.EntityPositionAndRotation,
		s.EntityID,
		s.DeltaX,
		s.DeltaY,
		s.DeltaZ,
		s.Yaw,
		s.Pitch,
		s.OnGround,
	)
}
