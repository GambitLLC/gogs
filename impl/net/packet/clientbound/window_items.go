package clientbound

import (
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/packetids"
)

type WindowItems struct {
	WindowID pk.UByte
	Count    pk.Short
	SlotData slotData
}

func (s WindowItems) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.WindowItems, s.WindowID, s.Count, s.SlotData)
}

type slotData []pk.Slot

func (a slotData) Encode() []byte {
	var bs []byte
	for _, v := range a {
		bs = append(bs, v.Encode()...)
	}
	return bs
}
