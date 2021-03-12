package clientbound

import (
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/packetids"
)

type SetSlot struct {
	WindowID pk.Byte
	Slot     pk.Short
	SlotData pk.Slot
}

func (s SetSlot) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.SetSlot, s.WindowID, s.Slot, s.SlotData)
}
