package server

import (
	"github.com/panjf2000/gnet"
	"gogs/impl/data"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/serverbound"
	"math"
)

func (s *Server) handlePlayerBlockPlacement(conn gnet.Conn, pkt pk.Packet) (out []byte, err error) {
	in := serverbound.PlayerBlockPlacement{}
	if err = in.FromPacket(pkt); err != nil {
		return
	}

	location := in.Location
	newX := int(math.Floor(float64(location.X) + float64(in.CursorPositionX)))
	newY := int(math.Floor(float64(location.Y) + float64(in.CursorPositionY)))
	newZ := int(math.Floor(float64(location.Z) + float64(in.CursorPositionZ)))

	if in.CursorPositionX == 0 {
		newX -= 1
	}
	if in.CursorPositionY == 0 {
		newY -= 1
	}
	if in.CursorPositionZ == 0 {
		newZ -= 1
	}

	// TODO: determine block id from player inventory
	player := s.playerFromConn(conn)

	player.InventoryLock.Lock()
	defer player.InventoryLock.Unlock()

	itemID := data.NamespacedID("minecraft:item", int32(player.Inventory[player.HeldItem+36].ItemID))
	blockID := data.BlockStateID(itemID, nil)

	if blockID != 0 {
		player.Inventory[player.HeldItem+36].ItemCount -= 1
		if player.Inventory[player.HeldItem+36].ItemCount == 0 {
			player.Inventory[player.HeldItem+36].Present = false
			player.Inventory[player.HeldItem+36].ItemID = 0
		}
		s.world.SetBlock(newX, newY, newZ, blockID)

		out = clientbound.BlockChange{
			Location: pk.Position{
				X: int32(newX),
				Y: int32(newY),
				Z: int32(newZ),
			},
			BlockID: pk.VarInt(blockID),
		}.CreatePacket().Encode()

		s.playerMapMutex.RLock()
		for c := range s.playerMap.connToPlayer {
			// TODO: block change packet should only be sent to players if chunk is loaded
			_ = c.AsyncWrite(out)
		}
		s.playerMapMutex.RUnlock()
	}

	return nil, nil
}
