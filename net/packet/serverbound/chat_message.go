package serverbound

import pk "gogs/net/packet"

type ChatMessage struct {
	Message pk.String
}

func (s *ChatMessage) FromPacket(packet pk.Packet) error {
	return packet.Unmarshal(&s.Message)
}
