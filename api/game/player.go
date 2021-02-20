package game

import (
	"github.com/google/uuid"
)

type Player struct {
	UUID          uuid.UUID
	Name          string
	Position      Position
	Rotation      Rotation
	SpawnPosition Position
}

type Position struct {
	X float64
	Y float64
	Z float64
}

type Rotation struct {
	Yaw   float32
	Pitch float32
}
