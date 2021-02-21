package clientbound

import (
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/packetids"
)

type SpawnPlayer struct {
	EntityID   pk.VarInt
	PlayerUUID pk.UUID
	X          pk.Double
	Y          pk.Double
	Z          pk.Double
	Yaw        pk.Angle
	Pitch      pk.Angle
}

func (s SpawnPlayer) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.SpawnPlayer, s.EntityID, s.PlayerUUID, s.X, s.Y, s.Z, s.Yaw, s.Pitch)
}
