package clientbound

import (
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/packetids"
)

type TimeUpdate struct {
	WorldAge  pk.Long
	TimeOfDay pk.Long
}

func (p TimeUpdate) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.TimeUpdate, p.WorldAge, p.TimeOfDay)
}
