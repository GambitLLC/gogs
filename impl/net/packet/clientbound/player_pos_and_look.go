package clientbound

import (
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
