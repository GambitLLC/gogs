package ecs

import (
	"gogs/impl/data"
	pk "gogs/impl/net/packet"
)

type ItemEntity struct {
	BasicEntity
	PositionComponent
	VelocityComponent
	Slot pk.Slot
}

func NewItem() *ItemEntity {
	return &ItemEntity{
		BasicEntity:       NewEntity(data.ProtocolID("minecraft:entity_type", "minecraft:item")),
		PositionComponent: PositionComponent{},
		VelocityComponent: VelocityComponent{},
		Slot:              pk.Slot{},
	}
}
