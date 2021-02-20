package clientbound

import pk "gogs/impl/net/packet"

type HeldItemChange struct {
	Slot	pk.Byte
}

func (s HeldItemChange) CreatePacket() pk.Packet {
	// TODO: create packetid consts
	return pk.Marshal(0x3F, s.Slot)
}