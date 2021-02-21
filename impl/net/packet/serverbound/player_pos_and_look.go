package serverbound

import pk "gogs/impl/net/packet"

type PlayerPositionAndLook struct {
	X          pk.Double
	Y          pk.Double
	Z          pk.Double
	Yaw        pk.Float
	Pitch      pk.Float
	TeleportID pk.VarInt
}

func (s *PlayerPositionAndLook) FromPacket(packet *pk.Packet) error {
	return packet.Unmarshal(&s.X, &s.Y, &s.Z, &s.Yaw, &s.Pitch, &s.TeleportID)
}
