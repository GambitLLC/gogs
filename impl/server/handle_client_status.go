package server

import (
	"bytes"
	"fmt"
	"gogs/impl/net"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
)

func (s *Server) handleClientStatus(conn net.Conn, pkt pk.Packet) ([]byte, error) {
	var actionID pk.VarInt
	if err := pkt.Unmarshal(&actionID); err != nil {
		return nil, err
	}

	switch actionID {
	case 0: // Perform respawn
		player := s.playerFromConn(conn)
		player.Health = 20

		buf := bytes.Buffer{}

		// send respawn packet
		buf.Write(clientbound.Respawn{
			Dimension:        pk.NBT{V: clientbound.MinecraftOverworld},
			WorldName:        "world",
			HashedSeed:       0,
			Gamemode:         pk.UByte(player.GameMode),
			PreviousGamemode: pk.UByte(player.GameMode),
			IsDebug:          false,
			IsFlat:           true,
			CopyMetadata:     false,
		}.CreatePacket().Encode())

		// send inventory
		player.InventoryLock.RLock()
		buf.Write(clientbound.WindowItems{
			WindowID: 0,
			Count:    pk.Short(len(player.Inventory)),
			SlotData: player.Inventory,
		}.CreatePacket().Encode())
		player.InventoryLock.RUnlock()

		return buf.Bytes(), nil
	case 1: // Request stats
	default:
		return nil, fmt.Errorf("client status got invalid action id %d", actionID)
	}

	return nil, nil
}
