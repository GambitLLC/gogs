package clientbound

import pk "gogs/impl/net/packet"

type ChatMessage struct {
	JSONData pk.Chat
	Position pk.Byte
	Sender   pk.UUID
}

func (p ChatMessage) CreatePacket() pk.Packet {
	return pk.Marshal(0x0E, p.JSONData, p.Position, p.Sender)
}
