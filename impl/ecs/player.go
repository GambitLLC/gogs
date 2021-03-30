package ecs

import (
	"github.com/google/uuid"
	pk "gogs/impl/net/packet"
	"sync"
)

type Player struct {
	BasicEntity
	PositionComponent
	SpawnPosition PositionComponent
	VelocityComponent
	RotationComponent
	HealthComponent
	FoodComponent
	InventoryComponent
	ClientSettingsComponent

	UUID uuid.UUID
	Name string

	GameMode uint8

	HeldSlotLock sync.Mutex
	HeldSlot     pk.Slot // item held on the cursor
	HeldItem     uint8   // hot bar slot which is selected

	PaintingLock  sync.RWMutex
	PaintingSlots []uint8

	// per connection data
	ConnectionComponent
	KnownChunks ChunkSet
}

type ChunkSet map[int]map[int]struct{} // x,z map

func (s ChunkSet) Add(x int, z int) {
	if s[x] == nil {
		s[x] = make(map[int]struct{})
	}
	s[x][z] = struct{}{}
}

func (s ChunkSet) Remove(x int, z int) {
	if s[x] == nil {
		return
	}
	delete(s[x], z)
}

func (s ChunkSet) Contains(x int, z int) bool {
	if s[x] == nil {
		return false
	}
	_, exists := s[x][z]
	return exists
}
