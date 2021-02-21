package serverbound

import pk "gogs/impl/net/packet"

type PlayerPositionAndRotation struct {
	X        pk.Double
	Y        pk.Double
	Z        pk.Double
	Yaw      pk.Float
	Pitch    pk.Float
	OnGround pk.Boolean
}

func (s *PlayerPositionAndRotation) FromPacket(packet *pk.Packet) error {
	return packet.Unmarshal(&s.X, &s.Y, &s.Z, &s.Yaw, &s.Pitch, &s.OnGround)
}
