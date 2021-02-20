package clientbound

import pk "gogs/impl/net/packet"

type SpawnPosition struct {
	Location	pk.Position
}

func (s SpawnPosition) CreatePacket() pk.Packet {
	// TODO: create packetid consts
	return pk.Marshal(0x42, s.Location)
}