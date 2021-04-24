package clientbound

import (
	pk "github.com/GambitLLC/gogs/net/packet"
	"github.com/GambitLLC/gogs/net/packet/packetids"
)

type UnloadChunk struct {
	ChunkX pk.Int
	ChunkZ pk.Int
}

func (s UnloadChunk) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.UnloadChunk, s.ChunkX, s.ChunkZ)
}
