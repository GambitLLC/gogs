package clientbound

import (
	pk "github.com/GambitLLC/gogs/net/packet"
	"github.com/GambitLLC/gogs/net/packet/packetids"
)

type HeldItemChange struct {
	Slot pk.Byte
}

func (s HeldItemChange) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.HeldItemChangeClientbound, s.Slot)
}
