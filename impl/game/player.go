package game

import (
	"github.com/google/uuid"
	"gogs/api"
	"gogs/api/data"
	"gogs/api/game"
)

type Player struct {
	game.Player
	EntityID      int32
	UUID          uuid.UUID
	Name          string
	Position      data.Position
	Rotation      data.Rotation
	SpawnPosition data.Position
}

func NewPlayer(name string, u uuid.UUID, entityID int32) *Player {
	spawnPos := data.Position{
		X: 0,
		Y: 1,
		Z: 0,
	}
	return &Player{
		EntityID: entityID,
		UUID:     u,
		Name:     name,
		Position: spawnPos,
		Rotation: data.Rotation{
			Yaw:   0,
			Pitch: 0,
		},
		SpawnPosition: spawnPos,
	}
}

func (p *Player) Tick(_ api.Server) {
	return
}

func (p Player) GetEntityID() int32 {
	return p.EntityID
}

func (p Player) GetUUID() uuid.UUID {
	return p.UUID
}

func (p Player) GetName() string {
	return p.Name
}

func (p Player) GetPosition() data.Position {
	return p.Position
}

func (p Player) GetRotation() data.Rotation {
	return p.Rotation
}

func (p *Player) SetPosition(position data.Position) {
	p.Position = position
}

func (p *Player) SetRotation(rotation data.Rotation) {
	p.Rotation = rotation
}

func (p Player) GetSpawnPosition() data.Position {
	return p.SpawnPosition
}
