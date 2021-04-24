package serverbound

import pk "github.com/GambitLLC/gogs/net/packet"

type PlayerPosition struct {
	X        pk.Double
	Y        pk.Double
	Z        pk.Double
	OnGround pk.Boolean
}

func (s *PlayerPosition) FromPacket(packet pk.Packet) error {
	return packet.Unmarshal(&s.X, &s.Y, &s.Z, &s.OnGround)
}
