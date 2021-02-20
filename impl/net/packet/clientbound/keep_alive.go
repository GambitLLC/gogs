package clientbound

import pk "gogs/impl/net/packet"

type KeepAlive struct {
	ID pk.Long
}

func (p KeepAlive) CreatePacket() pk.Packet {
	return pk.Marshal(0x1F, p.ID)
}
