package clientbound

import (
	pk "gogs/net/packet"
	"gogs/net/packet/packetids"
)

type HeldItemChange struct {
	Slot pk.Byte
}

func (s HeldItemChange) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.HeldItemChangeClientbound, s.Slot)
}
