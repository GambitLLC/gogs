package clientbound

import (
	pk "gogs/net/packet"
	"gogs/net/packet/packetids"
)

type EntityPosition struct {
	EntityID pk.VarInt
	DeltaX   pk.Short
	DeltaY   pk.Short
	DeltaZ   pk.Short
	OnGround pk.Boolean
}

func (s EntityPosition) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.EntityPosition, s.EntityID, s.DeltaX, s.DeltaY, s.DeltaZ, s.OnGround)
}
