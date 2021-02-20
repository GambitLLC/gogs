package clientbound

import pk "gogs/impl/net/packet"

type QueryStatusPong struct {
	Payload pk.Long
}

func (p QueryStatusPong) CreatePacket() pk.Packet {
	return pk.Marshal(0x01, p.Payload)
}
