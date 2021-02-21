package clientbound

import (
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/packetids"
)

type ChatMessage struct {
	JSONData pk.Chat
	Position pk.Byte
	Sender   pk.UUID
}

func (p ChatMessage) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.ChatMessageClientbound, p.JSONData, p.Position, p.Sender)
}
