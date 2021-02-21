package clientbound

import (
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/packetids"
)

type EntityHeadLook struct {
	EntityID pk.VarInt
	HeadYaw  pk.Angle
}

func (s EntityHeadLook) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.EntityHeadLook, s.EntityID, s.HeadYaw)
}
