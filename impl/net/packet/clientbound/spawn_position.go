package clientbound

import (
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/packetids"
)

type SpawnPosition struct {
	Location pk.Position
}

func (s SpawnPosition) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.SpawnPosition, s.Location)
}
