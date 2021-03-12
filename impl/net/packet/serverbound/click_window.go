package serverbound

import pk "gogs/impl/net/packet"

type ClickWindow struct {
	WindowID     pk.UByte
	Slot         pk.Short
	Button       pk.Byte
	ActionNumber pk.Short
	Mode         pk.VarInt
	ClickedItem  pk.Slot
}

func (s *ClickWindow) FromPacket(packet pk.Packet) error {
	return packet.Unmarshal(&s.WindowID, &s.Slot, &s.Button, &s.ActionNumber, &s.Mode, &s.ClickedItem)
}
