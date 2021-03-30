package server

import (
	"gogs/impl/ecs"
	"gogs/impl/logger"
	"gogs/impl/net"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/serverbound"
)

func (s *Server) handleClientSettings(conn net.Conn, pkt pk.Packet) (err error) {
	logger.Printf("Received client settings")
	in := serverbound.ClientSettings{}
	if err = in.FromPacket(pkt); err != nil {
		return
	}

	player := s.playerFromConn(conn)

	player.ChatMode = uint8(in.ChatMode)
	player.ViewDistance = byte(in.ViewDistance)
	if player.ViewDistance > s.ViewDistance {
		player.ViewDistance = s.ViewDistance
	}

	if !player.Online {
		player.Online = true
		err = s.sendInitialData(player)
	}

	return
}

// sendInitialData returns all the packets that a client typically needs when joining the game encoded
func (s *Server) sendInitialData(player *ecs.Player) (err error) {
	if err = player.Connection.WritePacket(clientbound.HeldItemChange{
		Slot: pk.Byte(player.HeldItem),
	}.CreatePacket()); err != nil {
		return
	}

	if err = player.Connection.WritePacket(clientbound.DeclareRecipes{
		NumRecipes: 0,
		Recipes:    nil,
	}.CreatePacket()); err != nil {
		return
	}

	if err = player.Connection.WritePacket(clientbound.VanillaTags().CreatePacket()); err != nil {
		return
	}

	if err = player.Connection.WritePacket(clientbound.PlayerPositionAndLook{
		X:          pk.Double(player.X),
		Y:          pk.Double(player.Y),
		Z:          pk.Double(player.Z),
		Yaw:        pk.Float(player.Yaw),
		Pitch:      pk.Float(player.Pitch),
		Flags:      0,
		TeleportID: 0,
	}.CreatePacket()); err != nil {
		return
	}

	if err = player.Connection.WritePacket(clientbound.PlayerPositionAndLook{
		X:          pk.Double(player.X),
		Y:          pk.Double(player.Y),
		Z:          pk.Double(player.Z),
		Yaw:        pk.Float(player.Yaw),
		Pitch:      pk.Float(player.Pitch),
		Flags:      0,
		TeleportID: 0,
	}.CreatePacket()); err != nil {
		return
	}

	// send player info (tab list)
	s.playerMap.Lock.RLock()
	defer s.playerMap.Lock.RUnlock()
	numPlayers := len(s.playerMap.connToPlayer)
	playerInfoArr := make([]pk.Encodable, 0, numPlayers)
	for _, p := range s.playerMap.connToPlayer {
		playerInfoArr = append(playerInfoArr, clientbound.PlayerInfoAddPlayer{
			UUID:           pk.UUID(p.UUID),
			Name:           pk.String(p.Name),
			NumProperties:  0,
			Properties:     nil,
			Gamemode:       pk.VarInt(p.GameMode),
			Ping:           1,
			HasDisplayName: false,
			DisplayName:    "",
		})
	}

	if err = player.Connection.WritePacket(clientbound.PlayerInfo{
		Action:     0,
		NumPlayers: pk.VarInt(numPlayers),
		Players:    playerInfoArr,
	}.CreatePacket()); err != nil {
		return
	}

	if err = player.Connection.WritePacket(clientbound.UpdateViewPosition{
		ChunkX: pk.VarInt(int32(player.X) >> 4),
		ChunkZ: pk.VarInt(int32(player.Z) >> 4),
	}.CreatePacket()); err != nil {
		return
	}

	// send initial chunks
	chunkX := int(player.X) >> 4
	chunkZ := int(player.Z) >> 4

	viewDistance := int(player.ViewDistance)
	for x := -viewDistance; x <= viewDistance; x++ {
		for z := -viewDistance; z <= viewDistance; z++ {
			if err = player.Connection.WritePacket(s.chunkDataPacket(chunkX+x, chunkZ+z)); err != nil {
				return
			}
		}
	}

	if err = player.Connection.WritePacket(clientbound.SpawnPosition{Location: pk.Position{
		X: int32(player.SpawnPosition.X),
		Y: int32(player.SpawnPosition.Y),
		Z: int32(player.SpawnPosition.Z),
	}}.CreatePacket()); err != nil {
		return
	}

	if err = player.Connection.WritePacket(clientbound.PlayerPositionAndLook{
		X:          pk.Double(player.X),
		Y:          pk.Double(player.Y),
		Z:          pk.Double(player.Z),
		Yaw:        pk.Float(player.Yaw),
		Pitch:      pk.Float(player.Pitch),
		Flags:      0,
		TeleportID: 0,
	}.CreatePacket()); err != nil {
		return
	}

	// send inventory
	player.InventoryLock.RLock()
	if err = player.Connection.WritePacket(clientbound.WindowItems{
		WindowID: 0,
		Count:    pk.Short(len(player.Inventory)),
		SlotData: player.Inventory,
	}.CreatePacket()); err != nil {
		return
	}
	player.InventoryLock.RUnlock()

	// send time update with negative time to keep sun in position
	if err = player.Connection.WritePacket(clientbound.TimeUpdate{
		WorldAge:  0,
		TimeOfDay: -6000,
	}.CreatePacket()); err != nil {
		return
	}

	// also add spawn player packets for players already online
	// TODO: this logic should be done elsewhere (when players enter range) (tick?)
	for _, p := range s.playerMap.connToPlayer {
		if p.UUID != player.UUID {
			if err = player.Connection.WritePacket(clientbound.SpawnPlayer{
				EntityID:   pk.VarInt(p.ID()),
				PlayerUUID: pk.UUID(p.UUID),
				X:          pk.Double(p.X),
				Y:          pk.Double(p.Y),
				Z:          pk.Double(p.Z),
				Yaw:        pk.Angle(p.Yaw),
				Pitch:      pk.Angle(p.Pitch),
			}.CreatePacket()); err != nil {
				return
			}
		}
	}

	return nil
}

func (s *Server) chunkDataPacket(x int, z int) pk.Packet {
	column := s.world.Column(x, z)

	column.Lock.RLock()
	defer column.Lock.RUnlock()

	var chunkDataArray clientbound.ChunkDataArray
	chunkDataArray = make(clientbound.ChunkDataArray, 0, 16)

	bitMask := 0
	for _, section := range column.Sections {
		if section == nil {
			continue
		}
		bitMask |= 1 << section.Y

		palette := make([]pk.VarInt, len(section.Palette))
		for i, blockID := range section.Palette {
			palette[i] = pk.VarInt(blockID)
		}

		blockData := make([]pk.Long, len(section.BlockStates.Data))
		for i, blockState := range section.BlockStates.Data {
			blockData[i] = pk.Long(blockState)
		}
		chunkDataArray = append(chunkDataArray, clientbound.ChunkSection{
			BlockCount:   4096,
			BitsPerBlock: pk.UByte(section.BlockStates.BitsPerValue),
			Palette: clientbound.ChunkPalette{
				Length:  pk.VarInt(len(palette)),
				Palette: palette,
			},
			DataArrayLength: pk.VarInt(len(blockData)),
			DataArray:       blockData,
		})
	}

	blockEntities := make([]pk.NBT, len(column.BlockEntities))
	for i, v := range column.BlockEntities {
		blockEntities[i] = pk.NBT{V: v}
	}

	return clientbound.ChunkData{
		ChunkX:         pk.Int(x),
		ChunkZ:         pk.Int(z),
		FullChunk:      true,
		PrimaryBitMask: pk.VarInt(bitMask),
		Heightmaps: pk.NBT{
			V: clientbound.Heightmap{
				MotionBlocking: make([]int64, 37),
				//WorldSurface:   motion_blocking,
			},
		},
		BiomesLength:     1024,
		Biomes:           make([]pk.VarInt, 1024),
		Size:             pk.VarInt(len(chunkDataArray.Encode())),
		Data:             chunkDataArray,
		NumBlockEntities: pk.VarInt(len(blockEntities)),
		BlockEntities:    blockEntities,
	}.CreatePacket()
}
