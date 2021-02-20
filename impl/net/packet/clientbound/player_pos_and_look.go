package clientbound

import pk "gogs/impl/net/packet"

type PlayerPositionAndLook struct {
	X			pk.Double
	Y			pk.Double
	Z			pk.Double
	Yaw			pk.Float
	Pitch		pk.Float
	Flags		pk.Byte
	TeleportID	pk.VarInt
}

func (s PlayerPositionAndLook) CreatePacket() pk.Packet {
	// TODO: create packetid consts
	return pk.Marshal(0x34, s.X, s.Y, s.Z, s.Yaw, s.Pitch, s.Flags, s.TeleportID)
}