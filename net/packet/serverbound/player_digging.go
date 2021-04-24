package serverbound

import pk "github.com/GambitLLC/gogs/net/packet"

type PlayerDigging struct {
	Status   pk.VarInt
	Location pk.Position
	Face     pk.Byte
}

func (s *PlayerDigging) FromPacket(packet pk.Packet) error {
	return packet.Unmarshal(&s.Status, &s.Location, &s.Face)
}
