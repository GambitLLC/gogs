package server

import (
	"github.com/GambitLLC/gogs/entities"
	pk "github.com/GambitLLC/gogs/net/packet"
	"github.com/GambitLLC/gogs/net/packet/clientbound"
)

// todo: put this into some entity manager
func (s *Server) spawnItem(item pk.Slot, location entities.PositionComponent) {
	itemEntity := entities.NewItem()
	itemEntity.PositionComponent = location
	itemEntity.Slot = item

	s.entityMap[itemEntity.ID()] = itemEntity

	s.broadcastPacket(clientbound.SpawnEntity{
		EntityID:   pk.VarInt(itemEntity.ID()),
		ObjectUUID: pk.UUID{},
		Type:       37, // item
		X:          pk.Double(location.X),
		Y:          pk.Double(location.Y),
		Z:          pk.Double(location.Z),
		Pitch:      0,
		Yaw:        0,
		Data:       1,
		VelocityX:  pk.Short(itemEntity.DeltaX),
		VelocityY:  pk.Short(itemEntity.DeltaY),
		VelocityZ:  pk.Short(itemEntity.DeltaZ),
	}.CreatePacket(), nil)

	s.broadcastPacket(clientbound.EntityMetadata{
		EntityID: pk.VarInt(itemEntity.ID()),
		Metadata: []clientbound.MetadataField{
			{Index: 7, Type: 6, Value: item}, // ITEM
			{Index: 0xFF},
		},
	}.CreatePacket(), nil)
}
