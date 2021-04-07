package clientbound

import (
	pk "gogs/net/packet"
	"gogs/net/packet/packetids"
)

type KeepAlive struct {
	ID pk.Long
}

func (p KeepAlive) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.KeepAliveClientbound, p.ID)
}
