package clientbound

import (
	"bytes"
	pk "gogs/impl/net/packet"
)

type ChunkData struct {
	ChunkX           pk.Int
	ChunkZ           pk.Int
	FullChunk        pk.Boolean
	PrimaryBitMask   pk.VarInt
	Heightmaps       pk.NBT
	BiomesLength     pk.VarInt   // Optional, not present if full chunk is false
	Biomes           BiomesArray // Optional, not present if full chunk is false
	Size             pk.VarInt
	Data             ChunkDataArray
	NumBlockEntities pk.VarInt
	BlockEntities    BlockEntities
}

func (s ChunkData) CreatePacket() pk.Packet {
	// TODO: create packetid consts
	if s.FullChunk {
		return pk.Marshal(0x20, s.ChunkX, s.ChunkZ, s.FullChunk, s.PrimaryBitMask,
			s.Heightmaps, s.BiomesLength, s.Biomes, s.Size, s.Data, s.NumBlockEntities,
			s.BlockEntities)
	} else {
		return pk.Marshal(0x20, s.ChunkX, s.ChunkZ, s.FullChunk, s.PrimaryBitMask,
			s.Heightmaps, s.Size, s.Data, s.NumBlockEntities, s.BlockEntities)
	}
}

type Heightmap struct {
	MotionBlocking []int64 `nbt:"MOTION_BLOCKING"`
	WorldSurface   []int64 `nbt:"WORLD_SURFACE"`
}

type BiomesArray []pk.VarInt

func (a BiomesArray) Encode() []byte {
	var bs []byte
	for _, v := range a {
		bs = append(bs, v.Encode()...)
	}
	return bs
}

type ChunkDataArray []ChunkSection

func (a ChunkDataArray) Encode() []byte {
	var bs []byte
	for _, v := range a {
		bs = append(bs, v.Encode()...)
	}
	return bs
}

type ChunkSection struct {
	BlockCount      pk.Short
	BitsPerBlock    pk.UByte
	Palette         ChunkPalette
	DataArrayLength pk.VarInt
	DataArray       []pk.Long
}

func (s ChunkSection) Encode() []byte {
	buf := bytes.Buffer{}
	buf.Write(s.BlockCount.Encode())
	buf.Write(s.BitsPerBlock.Encode())
	buf.Write(s.Palette.Encode())
	buf.Write(s.DataArrayLength.Encode())
	for _, v := range s.DataArray {
		buf.Write(v.Encode())
	}
	return buf.Bytes()
}

type ChunkPalette struct {
	Length  pk.VarInt
	Palette []pk.VarInt
}

func (s ChunkPalette) Encode() []byte {
	buf := bytes.Buffer{}
	buf.Write(s.Length.Encode())
	for _, v := range s.Palette {
		buf.Write(v.Encode())
	}
	return buf.Bytes()
}

type BlockEntities []pk.NBT

func (a BlockEntities) Encode() []byte {
	return nil
}
