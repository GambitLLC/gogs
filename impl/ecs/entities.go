package ecs

import (
	"github.com/google/uuid"
	pk "gogs/impl/net/packet"
	"sync/atomic"
)

var idCounter uint64

type BasicEntity struct {
	id uint64
}

func NewEntity() BasicEntity {
	return BasicEntity{id: atomic.AddUint64(&idCounter, 1)}
}

func (s BasicEntity) ID() uint64 {
	return s.id
}

type Player struct {
	BasicEntity
	PositionComponent
	VelocityComponent
	RotationComponent
	HealthComponent
	FoodComponent
	InventoryComponent
	ConnectionComponent
	ClientSettingsComponent

	SpawnPosition PositionComponent
	UUID          uuid.UUID
	Name          string

	HeldSlot pk.Slot // item held on the cursor
	HeldItem uint8   // hot bar slot which is selected
}
