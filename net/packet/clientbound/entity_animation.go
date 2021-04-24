package clientbound

import (
	pk "github.com/GambitLLC/gogs/net/packet"
	"github.com/GambitLLC/gogs/net/packet/packetids"
)

type EntityAnimation struct {
	EntityID  pk.VarInt
	Animation pk.UByte
}

func (s EntityAnimation) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.EntityAnimation, s.EntityID, s.Animation)
}
