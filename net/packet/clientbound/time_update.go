package clientbound

import (
	pk "github.com/GambitLLC/gogs/net/packet"
	"github.com/GambitLLC/gogs/net/packet/packetids"
)

type TimeUpdate struct {
	WorldAge  pk.Long
	TimeOfDay pk.Long
}

func (p TimeUpdate) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.TimeUpdate, p.WorldAge, p.TimeOfDay)
}
