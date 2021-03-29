package server

import (
	"fmt"
	"gogs/impl/net"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
)

func (s *Server) handleClientStatus(conn net.Conn, pkt pk.Packet) (err error) {
	var actionID pk.VarInt
	if err = pkt.Unmarshal(&actionID); err != nil {
		return
	}

	switch actionID {
	case 0: // Perform respawn
		player := s.playerFromConn(conn)
		player.Health = 20

		// send respawn packet
		if err = conn.WritePacket(clientbound.Respawn{
			Dimension:        pk.NBT{V: clientbound.MinecraftOverworld},
			WorldName:        "world",
			HashedSeed:       0,
			Gamemode:         pk.UByte(player.GameMode),
			PreviousGamemode: pk.UByte(player.GameMode),
			IsDebug:          false,
			IsFlat:           true,
			CopyMetadata:     false,
		}.CreatePacket()); err != nil {
			return
		}

		// send inventory
		player.InventoryLock.RLock()
		defer player.InventoryLock.RUnlock()
		if err = conn.WritePacket(clientbound.WindowItems{
			WindowID: 0,
			Count:    pk.Short(len(player.Inventory)),
			SlotData: player.Inventory,
		}.CreatePacket()); err != nil {
			return
		}
	case 1: // Request stats
	default:
		return fmt.Errorf("client status got invalid action id %d", actionID)
	}

	return nil
}
