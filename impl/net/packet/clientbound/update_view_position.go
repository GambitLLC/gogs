package clientbound

import pk "gogs/impl/net/packet"

type UpdateViewPosition struct {
	ChunkX pk.VarInt
	ChunkZ pk.VarInt
}

func (s UpdateViewPosition) CreatePacket() pk.Packet {
	// TODO: create packetid consts
	return pk.Marshal(0x40, s.ChunkX, s.ChunkZ)
}
