package clientbound

import (
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/packetids"
)

type BlockChange struct {
	Location pk.Position
	BlockID  pk.VarInt
}

func (p BlockChange) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.BlockChange, p.Location, p.BlockID)
}
