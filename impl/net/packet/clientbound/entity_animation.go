package clientbound

import (
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/packetids"
)

type EntityAnimation struct {
	EntityID  pk.VarInt
	Animation pk.UByte
}

func (s EntityAnimation) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.EntityAnimation, s.EntityID, s.Animation)
}
