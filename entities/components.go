package entities

import (
	"github.com/GambitLLC/gogs/net"
	pk "github.com/GambitLLC/gogs/net/packet"
	"sync"
)

type PositionComponent struct {
	X float64
	Y float64
	Z float64
}

type VelocityComponent struct {
	DeltaX int16 // In units of 1/8000 of a block per tick (50 ms)
	DeltaY int16
	DeltaZ int16
}

type RotationComponent struct {
	Pitch float32
	Yaw   float32
}

type HealthComponent struct {
	Health uint8
}

type FoodComponent struct {
	Food       uint8
	Saturation uint8
}

type InventoryComponent struct {
	InventoryLock sync.RWMutex
	Inventory     []pk.Slot
}

type ConnectionComponent struct {
	Connection net.Conn
	Online     bool
}

type ClientSettingsComponent struct {
	ViewDistance byte
	ChatMode     uint8
}
