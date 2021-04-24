package clientbound

import (
	pk "github.com/GambitLLC/gogs/net/packet"
	"github.com/GambitLLC/gogs/net/packet/packetids"
)

// https://wiki.vg/Protocol#Combat_Event
// Since only entity dead event is used, other fields are ignored and assumed
type CombatEvent struct {
	PlayerID pk.VarInt
	EntityID pk.Int
	Message  pk.Chat
}

func (s CombatEvent) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.CombatEvent, pk.VarInt(2), s.PlayerID, s.EntityID, s.Message)
}
