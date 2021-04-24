package clientbound

import (
	pk "github.com/GambitLLC/gogs/net/packet"
	"github.com/GambitLLC/gogs/net/packet/packetids"
)

type SetSlot struct {
	WindowID pk.Byte
	Slot     pk.Short
	SlotData pk.Slot
}

func (s SetSlot) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.SetSlot, s.WindowID, s.Slot, s.SlotData)
}
