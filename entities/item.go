package entities

import (
	"github.com/GambitLLC/gogs/data"
	pk "github.com/GambitLLC/gogs/net/packet"
)

type Item struct {
	BasicEntity
	PositionComponent
	VelocityComponent
	Slot pk.Slot
}

func NewItem() *Item {
	return &Item{
		BasicEntity:       NewEntity(data.ProtocolID("minecraft:entity_type", "minecraft:item")),
		PositionComponent: PositionComponent{},
		VelocityComponent: VelocityComponent{},
		Slot:              pk.Slot{},
	}
}
