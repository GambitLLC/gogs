package clientbound

import (
	"bytes"
	pk "gogs/impl/net/packet"
)

type ChunkData struct {
	ChunkX           pk.VarInt
	ChunkZ           pk.VarInt
	FullChunk        pk.Boolean
	PrimaryBitMask   pk.VarInt
	Heightmaps       pk.NBT
	BiomesLength     pk.VarInt		// Optional, not present if full chunk is false
	Biomes           BiomesArray	// Optional, not present if full chunk is false
	Size             pk.VarInt
	Data             DataArray
	NumBlockEntities pk.VarInt
	BlockEntities    BlockEntities
}

func (s ChunkData) Encode() []byte {
	buf := bytes.Buffer{}
	buf.Write(s.ChunkX.Encode())
	buf.Write(s.ChunkZ.Encode())
	buf.Write(s.FullChunk.Encode())
	buf.Write(s.PrimaryBitMask.Encode())
	buf.Write(s.Heightmaps.Encode())
	if s.FullChunk {
		// TODO: write biomes
	}
	buf.Write(s.Size.Encode())
	buf.Write(s.Data.Encode())
	buf.Write(s.NumBlockEntities.Encode())
	buf.Write(s.BlockEntities.Encode())

	return buf.Bytes()
}

type BiomesArray []pk.Encodable

func (a BiomesArray) Encode() []byte {
	return nil
}

type DataArray []pk.Encodable

func (a DataArray) Encode() []byte {
	return nil
}

type BlockEntities []pk.NBT

func (a BlockEntities) Encode() []byte {
	return nil
}