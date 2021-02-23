package serverbound

import pk "gogs/impl/net/packet"

type ClientSettings struct {
	Locale             pk.String
	ViewDistance       pk.Byte
	ChatMode           pk.VarInt
	ChatColors         pk.Boolean
	DisplayedSkinParts pk.UByte
	MainHand           pk.VarInt
}

func (s *ClientSettings) FromPacket(packet pk.Packet) error {
	return packet.Unmarshal(&s.Locale, &s.ViewDistance, &s.ChatMode, &s.ChatColors, &s.DisplayedSkinParts, &s.MainHand)
}
