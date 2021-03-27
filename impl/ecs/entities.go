package ecs

import (
	"github.com/google/uuid"
	pk "gogs/impl/net/packet"
	"sync"
	"sync/atomic"
)

var idCounter uint64

type BasicEntity struct {
	id         uint64
	entityType int32
}

func NewEntity(entityType int32) BasicEntity {
	return BasicEntity{
		id:         atomic.AddUint64(&idCounter, 1),
		entityType: entityType,
	}
}

func (s BasicEntity) ID() uint64 {
	return s.id
}

func (s BasicEntity) EntityType() int32 {
	return s.entityType
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

	GameMode uint8

	HeldSlotLock sync.Mutex
	HeldSlot     pk.Slot // item held on the cursor
	HeldItem     uint8   // hot bar slot which is selected

	PaintingLock  sync.RWMutex
	PaintingSlots []uint8
}

type ItemEntity struct {
	BasicEntity
	PositionComponent
	VelocityComponent
	Item pk.Slot
}
