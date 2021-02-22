package game

import (
	"github.com/google/uuid"
	"gogs/api/data"
)

type Player interface {
	EntityID() int32
	UUID() uuid.UUID
	Name() string
	Position() *data.Position
	Rotation() *data.Rotation
	SpawnPosition() *data.Position
}
