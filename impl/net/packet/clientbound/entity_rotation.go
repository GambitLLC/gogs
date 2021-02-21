package clientbound

import (
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/packetids"
)

type EntityRotation struct {
	EntityID pk.VarInt
	Yaw      pk.Angle
	Pitch    pk.Angle
	OnGround pk.Boolean
}

func (s EntityRotation) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.EntityRotation, s.EntityID, s.Yaw, s.Pitch, s.OnGround)
}
