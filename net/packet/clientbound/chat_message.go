package clientbound

import (
	pk "github.com/GambitLLC/gogs/net/packet"
	"github.com/GambitLLC/gogs/net/packet/packetids"
)

type ChatMessage struct {
	JSONData pk.Chat
	Position pk.Byte
	Sender   pk.UUID
}

func (p ChatMessage) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.ChatMessageClientbound, p.JSONData, p.Position, p.Sender)
}
