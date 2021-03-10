package ecs

import "github.com/panjf2000/gnet"

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
	Pitch uint8 // Rotation angle in steps of 1/256 of a full turn
	Yaw   uint8 // TODO: consider saving as float64 instead of following packet format...
}

type HealthComponent struct {
	Health uint8
}

type FoodComponent struct {
	Food       uint8
	Saturation uint8
}

type ConnectionComponent struct {
	Connection gnet.Conn
	Online     bool
}
