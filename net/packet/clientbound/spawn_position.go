package clientbound

import (
	pk "github.com/GambitLLC/gogs/net/packet"
	"github.com/GambitLLC/gogs/net/packet/packetids"
)

type SpawnPosition struct {
	Location pk.Position
}

func (s SpawnPosition) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.SpawnPosition, s.Location)
}
