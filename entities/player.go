package entities

import (
	"github.com/GambitLLC/gogs/data"
	pk "github.com/GambitLLC/gogs/net/packet"
	"github.com/google/uuid"
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

func NewPlayer() *Player {
	// zero values are not explicitly written
	return &Player{
		BasicEntity: NewEntity(data.ProtocolID("minecraft:entity_type", "minecraft:player")),
		PositionComponent: PositionComponent{
			X: 0,
			Y: 90,
			Z: 0,
		},
		SpawnPosition: PositionComponent{
			X: 0,
			Y: 90,
			Z: 0,
		},
		HealthComponent: HealthComponent{Health: 20},
		FoodComponent:   FoodComponent{Food: 20, Saturation: 0},
		InventoryComponent: InventoryComponent{
			Inventory: make([]pk.Slot, 46), // https://wiki.vg/Inventory#Player_Inventory,
		},
	}
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
