package clientbound

import (
	pk "gogs/net/packet"
	"gogs/net/packet/packetids"
)

type UnloadChunk struct {
	ChunkX pk.Int
	ChunkZ pk.Int
}

func (s UnloadChunk) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.UnloadChunk, s.ChunkX, s.ChunkZ)
}
