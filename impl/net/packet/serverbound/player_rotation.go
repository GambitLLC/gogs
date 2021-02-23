package serverbound

import pk "gogs/impl/net/packet"

type PlayerRotation struct {
	Yaw      pk.Float
	Pitch    pk.Float
	OnGround pk.Boolean
}

func (s *PlayerRotation) FromPacket(packet pk.Packet) error {
	return packet.Unmarshal(&s.Yaw, &s.Pitch, &s.OnGround)
}
