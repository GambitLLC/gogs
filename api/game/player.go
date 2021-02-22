package game

import (
	"github.com/google/uuid"
	"gogs/api/data"
)

type Player interface {
	GetEntityID() int32
	GetUUID() uuid.UUID
	GetName() string
	GetPosition() data.Position
	GetRotation() data.Rotation
	GetSpawnPosition() data.Position
	SetPosition(data.Position)
	SetRotation(data.Rotation)
}
