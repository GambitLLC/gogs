package clientbound

import (
	pk "github.com/GambitLLC/gogs/net/packet"
	"github.com/GambitLLC/gogs/net/packet/packetids"
)

type KeepAlive struct {
	ID pk.Long
}

func (p KeepAlive) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.KeepAliveClientbound, p.ID)
}
