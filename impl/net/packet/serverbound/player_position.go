package serverbound

import pk "gogs/impl/net/packet"

type PlayerPosition struct {
	X        pk.Double
	Y        pk.Double
	Z        pk.Double
	OnGround pk.Boolean
}

func (s *PlayerPosition) FromPacket(packet pk.Packet) error {
	return packet.Unmarshal(&s.X, &s.Y, &s.Z, &s.OnGround)
}
