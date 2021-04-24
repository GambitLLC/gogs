package clientbound

import (
	pk "github.com/GambitLLC/gogs/net/packet"
	"github.com/GambitLLC/gogs/net/packet/packetids"
)

type StatusPong struct {
	Payload pk.Long
}

func (p StatusPong) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.StatusPong, p.Payload)
}
