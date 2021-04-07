package clientbound

import (
	pk "gogs/net/packet"
	"gogs/net/packet/packetids"
)

type UpdateViewPosition struct {
	ChunkX pk.VarInt
	ChunkZ pk.VarInt
}

func (s UpdateViewPosition) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.UpdateViewPosition, s.ChunkX, s.ChunkZ)
}
