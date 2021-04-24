package clientbound

import (
	pk "github.com/GambitLLC/gogs/net/packet"
	"github.com/GambitLLC/gogs/net/packet/packetids"
)

type EntityHeadLook struct {
	EntityID pk.VarInt
	HeadYaw  pk.Angle
}

func (s EntityHeadLook) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.EntityHeadLook, s.EntityID, s.HeadYaw)
}
