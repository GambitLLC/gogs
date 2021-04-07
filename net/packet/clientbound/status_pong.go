package clientbound

import (
	pk "gogs/net/packet"
	"gogs/net/packet/packetids"
)

type StatusPong struct {
	Payload pk.Long
}

func (p StatusPong) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.StatusPong, p.Payload)
}
