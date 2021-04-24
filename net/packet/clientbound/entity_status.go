package clientbound

import (
	pk "github.com/GambitLLC/gogs/net/packet"
	"github.com/GambitLLC/gogs/net/packet/packetids"
)

type EntityStatus struct {
	EntityID     pk.Int
	EntityStatus pk.Byte
}

// https://wiki.vg/Entity_statuses

func (s EntityStatus) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.EntityStatus, s.EntityID, s.EntityStatus)
}
