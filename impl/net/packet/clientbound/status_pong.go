package clientbound

import (
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/packetids"
)

type StatusPong struct {
	Payload pk.Long
}

func (p StatusPong) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.StatusPong, p.Payload)
}
