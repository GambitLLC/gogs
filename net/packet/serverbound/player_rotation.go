package serverbound

import pk "github.com/GambitLLC/gogs/net/packet"

type PlayerRotation struct {
	Yaw      pk.Float
	Pitch    pk.Float
	OnGround pk.Boolean
}

func (s *PlayerRotation) FromPacket(packet pk.Packet) error {
	return packet.Unmarshal(&s.Yaw, &s.Pitch, &s.OnGround)
}
