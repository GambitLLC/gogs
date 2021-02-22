package clientbound

import (
	"gogs/api/game"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/packetids"
)

type PlayerPositionAndLook struct {
	X          pk.Double
	Y          pk.Double
	Z          pk.Double
	Yaw        pk.Float
	Pitch      pk.Float
	Flags      pk.Byte
	TeleportID pk.VarInt
}

func (s PlayerPositionAndLook) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.PlayerPositionAndLookClientbound, s.X, s.Y, s.Z, s.Yaw, s.Pitch, s.Flags, s.TeleportID)
}

func (s *PlayerPositionAndLook) FromPlayer(p game.Player) *PlayerPositionAndLook {
	pos := p.GetPosition()
	rot := p.GetRotation()
	s.X = pk.Double(pos.X)
	s.Y = pk.Double(pos.Y)
	s.Z = pk.Double(pos.Z)
	s.Yaw = pk.Float(rot.Yaw)
	s.Pitch = pk.Float(rot.Pitch)
	return s
}
