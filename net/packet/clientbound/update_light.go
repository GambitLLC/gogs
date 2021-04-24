package clientbound

import (
	"bytes"
	pk "github.com/GambitLLC/gogs/net/packet"
	"github.com/GambitLLC/gogs/net/packet/packetids"
)

type UpdateLight struct {
	ChunkX              pk.VarInt
	ChunkZ              pk.VarInt
	TrustEdges          pk.Boolean
	SkyLightMask        pk.VarInt
	BlockLightMask      pk.VarInt
	EmptySkyLightMask   pk.VarInt
	EmptyBlockLightMask pk.VarInt
	SkyLightArrays      SkyLight
	BlockLightArrays    BlockLight
}

func (s UpdateLight) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.UpdateLight, s.ChunkX, s.ChunkZ, s.TrustEdges, s.SkyLightMask,
		s.BlockLightMask, s.EmptySkyLightMask, s.EmptyBlockLightMask,
		s.SkyLightArrays, s.BlockLightArrays)
}

type SkyLight struct {
	Arrays []pk.ByteArray
}

func (s SkyLight) Encode() []byte {
	buf := bytes.Buffer{}
	for _, v := range s.Arrays {
		buf.Write(v.Encode())
	}
	return buf.Bytes()
}

type BlockLight struct {
	Arrays []pk.ByteArray
}

func (s BlockLight) Encode() []byte {
	buf := bytes.Buffer{}
	for _, v := range s.Arrays {
		buf.Write(v.Encode())
	}
	return buf.Bytes()
}
