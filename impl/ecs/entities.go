package ecs

import (
	"github.com/google/uuid"
	pk "gogs/impl/net/packet"
	"sync"
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

	GameMode uint8

	HeldSlotLock sync.Mutex
	HeldSlot     pk.Slot // item held on the cursor
	HeldItem     uint8   // hot bar slot which is selected

	PaintingLock  sync.RWMutex
	PaintingSlots []uint8

	chunkManager map[int]map[int]struct{}
}

// TrackChunk marks a chunk as loaded and returns whether or not it was previously loaded
func (p *Player) TrackChunk(x int, z int) bool {
	if p.chunkManager == nil {
		p.chunkManager = make(map[int]map[int]struct{})
	}

	if p.chunkManager[x] == nil {
		p.chunkManager[x] = make(map[int]struct{})
	}

	_, exists := p.chunkManager[x][z]
	if !exists {
		p.chunkManager[x][z] = struct{}{}
	}

	return exists
}

type ItemEntity struct {
	BasicEntity
	PositionComponent
	VelocityComponent
	Item pk.Slot
}
