package game

import (
	"github.com/google/uuid"
	"github.com/panjf2000/gnet"
	"gogs/api/data"
)

type Player struct {
	entityID      int32
	uuid          uuid.UUID
	name          string
	health        uint8
	position      data.Position
	rotation      data.Rotation
	spawnPosition data.Position
	conn          gnet.Conn
}

func NewPlayer(name string, u uuid.UUID, c gnet.Conn, entityID int32) *Player {
	spawnPos := data.Position{
		X: 0,
		Y: 1,
		Z: 0,
	}
	return &Player{
		entityID: entityID,
		uuid:     u,
		name:     name,
		health:   20,
		position: spawnPos,
		rotation: data.Rotation{
			Yaw:   0,
			Pitch: 0,
		},
		spawnPosition: spawnPos,
		conn:          c,
	}
}

func (p Player) EntityID() int32 {
	return p.entityID
}

func (p Player) UUID() uuid.UUID {
	return p.uuid
}

func (p Player) Name() string {
	return p.name
}

func (p *Player) Position() *data.Position {
	return &p.position
}

func (p *Player) Rotation() *data.Rotation {
	return &p.rotation
}

func (p *Player) SpawnPosition() *data.Position {
	return &p.spawnPosition
}

func (p Player) Conn() gnet.Conn {
	return p.conn
}

func (p *Player) Health() *uint8 {
	return &p.health
}
