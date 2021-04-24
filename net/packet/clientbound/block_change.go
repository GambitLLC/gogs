package clientbound

import (
	pk "github.com/GambitLLC/gogs/net/packet"
	"github.com/GambitLLC/gogs/net/packet/packetids"
)

type BlockChange struct {
	Location pk.Position
	BlockID  pk.VarInt
}

func (p BlockChange) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.BlockChange, p.Location, p.BlockID)
}
