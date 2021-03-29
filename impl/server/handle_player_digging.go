package server

import (
	"gogs/impl/data"
	"gogs/impl/ecs"
	"gogs/impl/net"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/serverbound"
)

func (s *Server) handlePlayerDigging(conn net.Conn, pkt pk.Packet) (err error) {
	in := serverbound.PlayerDigging{}
	if err = in.FromPacket(pkt); err != nil {
		return
	}

	player := s.playerFromConn(conn)

	switch in.Status {
	case 4: // Drop item
		player.InventoryLock.Lock()
		item := &player.Inventory[player.HeldItem+36]
		if item.Present {
			s.spawnItem(pk.Slot{
				Present:   true,
				ItemID:    item.ItemID,
				ItemCount: 1,
				NBT:       item.NBT,
			}, player.PositionComponent)

			item.ItemCount -= 1
			if item.ItemCount == 0 {
				item.Present = false
				item.ItemID = 0
			}
		}
		player.InventoryLock.Unlock()

	}

	return
}

func (s *Server) spawnItem(item pk.Slot, location ecs.PositionComponent) {
	itemEntity := ecs.ItemEntity{
		BasicEntity:       ecs.NewEntity(data.ProtocolID("minecraft:entity_type", "minecraft:item")),
		PositionComponent: location,
		VelocityComponent: ecs.VelocityComponent{},
		Item:              item,
	}

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
