package game

import (
	"github.com/google/uuid"
	"gogs/api"
	"gogs/api/data"
	"gogs/api/game"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/packetids"
)

type Player struct {
	game.Player
	EntityID        int32
	UUID            uuid.UUID
	Name            string
	CurrentPosition data.Position
	CurrentRotation data.Rotation
	NextPosition    data.Position
	NextRotation    data.Rotation
	SpawnPosition   data.Position
}

func NewPlayer(name string, u uuid.UUID, entityID int32) *Player {
	spawnPos := data.Position{
		X: 0,
		Y: 1,
		Z: 0,
	}
	return &Player{
		EntityID:        entityID,
		UUID:            u,
		Name:            name,
		CurrentPosition: spawnPos,
		CurrentRotation: data.Rotation{
			Yaw:   0,
			Pitch: 0,
		},
		SpawnPosition: spawnPos,
	}
}

func (p *Player) Tick(s api.Server) {
	// TODO: make better way to send out positions
	deltaPos := data.Position{
		X:        (p.NextPosition.X*32 - p.CurrentPosition.X*32) * 128,
		Y:        (p.NextPosition.Y*32 - p.CurrentPosition.Y*32) * 128,
		Z:        (p.NextPosition.Z*32 - p.CurrentPosition.Z*32) * 128,
		OnGround: p.NextPosition.OnGround || p.NextRotation.OnGround,
	}
	posChanged := deltaPos.X != 0 || deltaPos.Y != 0 || deltaPos.Z != 0

	if posChanged && p.NextRotation != p.CurrentRotation {
		outPacket := clientbound.EntityPositionAndRotation{
			EntityID: pk.VarInt(p.EntityID),
			DeltaX:   pk.Short(deltaPos.X),
			DeltaY:   pk.Short(deltaPos.Y),
			DeltaZ:   pk.Short(deltaPos.Z),
			Yaw:      pk.Angle(p.NextRotation.Yaw / 360 * 256),
			Pitch:    pk.Angle(p.NextRotation.Pitch / 360 * 256),
			OnGround: pk.Boolean(p.NextPosition.OnGround),
		}.CreatePacket().Encode()
		// also send head rotation packet
		outPacket = append(outPacket, clientbound.EntityHeadLook{
			EntityID: pk.VarInt(p.EntityID),
			HeadYaw:  pk.Angle(p.NextRotation.Yaw / 360 * 256),
		}.CreatePacket().Encode()...)

		for _, player := range s.Players() {
			conn := s.ConnFromUUID(player.GetUUID())
			if p.UUID.ID() != player.GetUUID().ID() {
				_ = conn.AsyncWrite(outPacket)
			}
		}
	} else if posChanged {
		// send new player position to everyone else
		outPacket := clientbound.EntityPosition{
			EntityID: pk.VarInt(p.EntityID),
			DeltaX:   pk.Short(deltaPos.X),
			DeltaY:   pk.Short(deltaPos.Y),
			DeltaZ:   pk.Short(deltaPos.Z),
			OnGround: pk.Boolean(p.NextPosition.OnGround),
		}.CreatePacket().Encode()
		for _, player := range s.Players() {
			conn := s.ConnFromUUID(player.GetUUID())
			if p.UUID.ID() != player.GetUUID().ID() {
				_ = conn.AsyncWrite(outPacket)
			}
		}
	} else if p.NextRotation != p.CurrentRotation {
		// send entity rotation packet
		outPacket := clientbound.EntityRotation{
			EntityID: pk.VarInt(p.EntityID),
			Yaw:      pk.Angle(p.NextRotation.Yaw / 360 * 256),
			Pitch:    pk.Angle(p.NextRotation.Pitch / 360 * 256),
			OnGround: pk.Boolean(p.NextRotation.OnGround),
		}.CreatePacket().Encode()

		// also send head rotation packet
		outPacket = append(outPacket, clientbound.EntityHeadLook{
			EntityID: pk.VarInt(p.EntityID),
			HeadYaw:  pk.Angle(p.NextRotation.Yaw / 360 * 256),
		}.CreatePacket().Encode()...)

		for _, player := range s.Players() {
			conn := s.ConnFromUUID(player.GetUUID())
			if p.UUID != player.GetUUID() {
				_ = conn.AsyncWrite(outPacket)
			}
		}
	} else {
		for _, player := range s.Players() {
			conn := s.ConnFromUUID(player.GetUUID())
			if p.UUID != player.GetUUID() {
				_ = conn.AsyncWrite(pk.Marshal(packetids.EntityMovement, pk.VarInt(p.EntityID)).Encode())
			}
		}
	}
	p.CurrentPosition = p.NextPosition
	p.CurrentRotation = p.NextRotation
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
	return p.CurrentPosition
}

func (p Player) GetRotation() data.Rotation {
	return p.CurrentRotation
}

func (p *Player) SetPosition(position data.Position) {
	p.NextPosition = position
}

func (p *Player) SetRotation(rotation data.Rotation) {
	p.NextRotation = rotation
}

func (p Player) GetSpawnPosition() data.Position {
	return p.SpawnPosition
}
