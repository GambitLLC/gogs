package serverbound

import pk "gogs/impl/net/packet"

type QueryStatusPing struct {
	Payload pk.Long
}

func (p QueryStatusPing) CreatePacket() pk.Packet {
	return pk.Marshal(0x01, p.Payload)
}
