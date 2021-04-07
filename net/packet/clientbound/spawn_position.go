package clientbound

import (
	pk "gogs/net/packet"
	"gogs/net/packet/packetids"
)

type SpawnPosition struct {
	Location pk.Position
}

func (s SpawnPosition) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.SpawnPosition, s.Location)
}
