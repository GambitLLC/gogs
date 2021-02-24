package clientbound

import (
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/packetids"
)

type EntityStatus struct {
	EntityID     pk.Int
	EntityStatus pk.Byte
}

// https://wiki.vg/Entity_statuses

func (s EntityStatus) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.EntityStatus, s.EntityID, s.EntityStatus)
}
