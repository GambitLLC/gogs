package clientbound

import (
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/packetids"
)

type KeepAlive struct {
	ID pk.Long
}

func (p KeepAlive) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.KeepAliveClientbound, p.ID)
}
