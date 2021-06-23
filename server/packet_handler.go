package server

import (
	"fmt"
	"math"

	"github.com/GambitLLC/gogs/chat"
	"github.com/GambitLLC/gogs/data"
	"github.com/GambitLLC/gogs/entities"
	"github.com/GambitLLC/gogs/logger"
	"github.com/GambitLLC/gogs/net"
	pk "github.com/GambitLLC/gogs/net/packet"
	"github.com/GambitLLC/gogs/net/packet/clientbound"
	"github.com/GambitLLC/gogs/net/packet/packetids"
	"github.com/GambitLLC/gogs/net/packet/serverbound"
)

func (s *Server) onLoginStart(conn net.Conn, pkt pk.Packet) error {
	var name pk.String
	if err := pkt.Unmarshal(&name); err != nil {
		return err
	}

	// TODO: handle encryption (send encryption start)

	u := pk.NameToUUID(string(name)) // todo: get uuid from mojang servers
	if err := conn.WritePacket(pk.Marshal(
		packetids.LoginSuccess,
		pk.UUID(u),
		name,
	)); err != nil {
		return err
	}

	player := s.createPlayer(string(name), u, conn)
	if err := conn.WritePacket(s.joinGamePacket(player)); err != nil {
		return err
	}

	s.Broadcast(fmt.Sprintf("%v has joined the game", name))

	// send out new player info to everyone already online
	s.broadcastPacket(clientbound.PlayerInfo{
		Action:     0,
		NumPlayers: 1,
		Players: []pk.Encodable{
			clientbound.PlayerInfoAddPlayer{
				UUID:           pk.UUID(player.UUID),
				Name:           pk.String(player.Name),
				NumProperties:  0,
				Properties:     nil,
				Gamemode:       pk.VarInt(player.GameMode),
				Ping:           1,
				HasDisplayName: false,
				DisplayName:    "",
			},
		},
	}.CreatePacket(), conn)

	// TODO: spawn player should be occurring when players enter range (not join game), do logic elsewhere (tick?)
	s.broadcastPacket(clientbound.SpawnPlayer{
		EntityID:   pk.VarInt(player.ID()),
		PlayerUUID: pk.UUID(player.UUID),
		X:          pk.Double(player.X),
		Y:          pk.Double(player.Y),
		Z:          pk.Double(player.Z),
		Yaw:        pk.Angle(player.Yaw / 360 * 256),
		Pitch:      pk.Angle(player.Pitch / 360 * 256),
	}.CreatePacket(), conn)

	return nil
}

func (s *Server) onClientSettings(conn net.Conn, pkt pk.Packet) error {
	in := serverbound.ClientSettings{}
	if err := in.FromPacket(pkt); err != nil {
		return err
	}

	player := s.playerFromConn(conn)

	player.ChatMode = uint8(in.ChatMode)
	player.ViewDistance = byte(in.ViewDistance)
	if player.ViewDistance > s.ViewDistance {
		player.ViewDistance = s.ViewDistance
	}

	if !player.Online {
		player.Online = true
		if err := s.sendInitialData(player); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) onChatMessage(conn net.Conn, pkt pk.Packet) error {
	m := serverbound.ChatMessage{}
	if err := m.FromPacket(pkt); err != nil {
		return err
	}

	player := s.playerFromConn(conn)
	logger.Printf("Received chat message `%v` from %v", m.Message, player.Name)

	// TODO: MOVE THIS INTO COMMAND HANDLER
	if m.Message == "/stop" {
		s.stop()
		return nil
	}

	msg := chat.NewTranslationComponent(
		"chat.type.text", // "<%s> %s"
		chat.NewStringComponent(player.Name),
		chat.NewStringComponent(string(m.Message)),
	)

	s.broadcastPacket(clientbound.ChatMessage{
		JSONData: pk.Chat(msg.AsJSON()),
		Position: chat.Chat,
		Sender:   pk.UUID(player.UUID),
	}.CreatePacket(), nil)

	return nil
}

func (s *Server) onPlayerPosition(conn net.Conn, pkt pk.Packet) error {
	player := s.playerFromConn(conn)

	in := serverbound.PlayerPosition{}
	if err := in.FromPacket(pkt); err != nil {
		return err
	}

	s.broadcastPacket(clientbound.EntityPosition{
		EntityID: pk.VarInt(player.ID()),
		DeltaX:   pk.Short((float64(in.X*32) - player.X*32) * 128),
		DeltaY:   pk.Short((float64(in.Y*32) - player.Y*32) * 128),
		DeltaZ:   pk.Short((float64(in.Z*32) - player.Z*32) * 128),
		OnGround: in.OnGround,
	}.CreatePacket(), conn)

	// TODO: according to wikivg, update view position is sent on change in Y coord as well
	// if chunk border was crossed, update view pos and send new chunks
	if int(player.X)>>4 != int(in.X)>>4 || int(player.Z)>>4 != int(in.Z)>>4 {
		if err := s.updateViewPosition(player); err != nil {
			return err
		}
	}

	// update player position
	player.X = float64(in.X)
	player.Y = float64(in.Y)
	player.Z = float64(in.Z)

	return nil
}

func (s *Server) onPlayerPositionAndRotation(conn net.Conn, pkt pk.Packet) error {
	player := s.playerFromConn(conn)
	in := serverbound.PlayerPositionAndRotation{}
	if err := in.FromPacket(pkt); err != nil {
		return err
	}

	s.broadcastPacket(clientbound.EntityPositionAndRotation{
		EntityID: pk.VarInt(player.ID()),
		DeltaX:   pk.Short((float64(in.X*32) - player.X*32) * 128),
		DeltaY:   pk.Short((float64(in.Y*32) - player.Y*32) * 128),
		DeltaZ:   pk.Short((float64(in.Z*32) - player.Z*32) * 128),
		Yaw:      pk.Angle(in.Yaw / 360 * 256),
		Pitch:    pk.Angle(in.Pitch / 360 * 256),
		OnGround: in.OnGround,
	}.CreatePacket(), conn)

	// also send head rotation packet
	s.broadcastPacket(clientbound.EntityHeadLook{
		EntityID: pk.VarInt(player.ID()),
		HeadYaw:  pk.Angle(in.Yaw / 360 * 256),
	}.CreatePacket(), conn)

	// if chunk border was crossed, update view pos and send new chunks
	if int(player.X)>>4 != int(in.X)>>4 || int(player.Z)>>4 != int(in.Z)>>4 {
		if err := s.updateViewPosition(player); err != nil {
			return err
		}
	}

	player.X = float64(in.X)
	player.Y = float64(in.Y)
	player.Z = float64(in.Z)

	player.Yaw = float32(in.Yaw)
	player.Pitch = float32(in.Pitch)

	return nil
}

func (s *Server) onPlayerRotation(conn net.Conn, pkt pk.Packet) error {
	player := s.playerFromConn(conn)
	in := serverbound.PlayerRotation{}
	if err := in.FromPacket(pkt); err != nil {
		return err
	}

	s.broadcastPacket(clientbound.EntityRotation{
		EntityID: pk.VarInt(player.ID()),
		Yaw:      pk.Angle(in.Yaw / 360 * 256),
		Pitch:    pk.Angle(in.Pitch / 360 * 256),
		OnGround: in.OnGround,
	}.CreatePacket(), conn)

	// also send head rotation packet
	s.broadcastPacket(clientbound.EntityHeadLook{
		EntityID: pk.VarInt(player.ID()),
		HeadYaw:  pk.Angle(in.Yaw / 360 * 256),
	}.CreatePacket(), conn)

	player.Yaw = float32(in.Yaw)
	player.Pitch = float32(in.Pitch)

	return nil
}

func (s *Server) onAnimation(conn net.Conn, pkt pk.Packet) error {
	var hand pk.VarInt
	if err := pkt.Unmarshal(&hand); err != nil {
		return err
	}

	if hand != 0 && hand != 1 {
		_ = conn.Close() // TODO: send disconnect packet
		if err := fmt.Errorf("animation handler got invalid hand %d", hand); err != nil {
			return err
		}
	}

	player := s.playerFromConn(conn)

	anim := 0 // swing main arm
	if hand == 1 {
		anim = 3 // swing off hand
	}

	s.broadcastPacket(clientbound.EntityAnimation{
		EntityID:  pk.VarInt(player.ID()),
		Animation: pk.UByte(anim),
	}.CreatePacket(), conn)

	return nil
}

func (s *Server) onEntityAction(conn net.Conn, pkt pk.Packet) error {
	in := serverbound.EntityAction{}
	if err := in.FromPacket(pkt); err != nil {
		return err
	}

	player := s.playerFromConn(conn)
	switch in.ActionID {
	case 0: // start sneaking
		s.broadcastPacket(clientbound.EntityMetadata{
			EntityID: pk.VarInt(player.ID()),
			Metadata: []clientbound.MetadataField{
				{Index: 6, Type: 18, Value: pk.VarInt(5)}, // SNEAKING
				{Index: 0xFF},
			},
		}.CreatePacket(), conn)
	case 1: // stop sneaking
		s.broadcastPacket(clientbound.EntityMetadata{
			EntityID: pk.VarInt(player.ID()),
			Metadata: []clientbound.MetadataField{
				{Index: 6, Type: 18, Value: pk.VarInt(0)}, // STANDING
				{Index: 0xFF},
			},
		}.CreatePacket(), conn)
	case 2: // leave bed
	case 3: // start sprinting
	case 4: // stop sprinting
	case 5: // start jump with horse
	case 6: // stop jump with horse
	case 7: // open horse inventory
	case 8: // start flying with elytra
	default:
		return fmt.Errorf("entity action got invalid action id %d", in.ActionID)
	}

	return nil
}

func (s *Server) onInteractEntity(conn net.Conn, pkt pk.Packet) error {
	in := serverbound.InteractEntity{}
	if err := in.FromPacket(pkt); err != nil {
		return err
	}

	logger.Printf("received interact entity")

	switch in.Type {
	case 0: // interact
	case 1: // attack
		player := s.entityFromID(uint64(in.EntityID)).(*entities.Player)
		if player == nil {
			if err := fmt.Errorf("interact entity could not find entity with id %d", in.EntityID); err != nil {
				return err
			}
		}
		if err := s.handleAttack(s.playerFromConn(conn), player); err != nil {
			return err
		}
	case 2: // interact at
	default:
		_ = conn.Close()
		if err := fmt.Errorf("interact entity got invalid type %d", in.Type); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) onClientStatus(conn net.Conn, pkt pk.Packet) error {
	var actionID pk.VarInt
	if err := pkt.Unmarshal(&actionID); err != nil {
		return err
	}

	switch actionID {
	case 0: // Perform respawn
		player := s.playerFromConn(conn)
		player.Health = 20
		player.PositionComponent = player.SpawnPosition

		// send respawn packet
		if err := conn.WritePacket(clientbound.Respawn{
			Dimension:        pk.NBT{V: clientbound.MinecraftOverworld},
			WorldName:        "world",
			HashedSeed:       0,
			Gamemode:         pk.UByte(player.GameMode),
			PreviousGamemode: pk.UByte(player.GameMode),
			IsDebug:          false,
			IsFlat:           true,
			CopyMetadata:     false,
		}.CreatePacket()); err != nil {
			return err
		}

		// send inventory
		player.InventoryLock.RLock()
		defer player.InventoryLock.RUnlock()
		if err := conn.WritePacket(clientbound.WindowItems{
			WindowID: 0,
			Count:    pk.Short(len(player.Inventory)),
			SlotData: player.Inventory,
		}.CreatePacket()); err != nil {
			return err
		}

		// spawn player for everyone else online
		s.broadcastPacket(clientbound.SpawnPlayer{
			EntityID:   pk.VarInt(player.ID()),
			PlayerUUID: pk.UUID(player.UUID),
			X:          pk.Double(player.X),
			Y:          pk.Double(player.Y),
			Z:          pk.Double(player.Z),
			Yaw:        pk.Angle(player.Yaw / 360 * 256),
			Pitch:      pk.Angle(player.Pitch / 360 * 256),
		}.CreatePacket(), conn)
	case 1: // Request stats
	default:
		return fmt.Errorf("client status got invalid action id %d", actionID)
	}

	return nil
}

func (s *Server) onPlayerDigging(conn net.Conn, pkt pk.Packet) error {
	in := serverbound.PlayerDigging{}
	if err := in.FromPacket(pkt); err != nil {
		return err
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

	return nil
}

func (s *Server) onPlayerBlockPlacement(conn net.Conn, pkt pk.Packet) error {
	in := serverbound.PlayerBlockPlacement{}
	if err := in.FromPacket(pkt); err != nil {
		return err
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

	player.InventoryLock.RLock()
	itemID := data.NamespacedID("minecraft:item", int32(player.Inventory[player.HeldItem+36].ItemID))
	player.InventoryLock.RUnlock()
	blockID := data.BlockStateID(itemID, nil)

	if blockID != 0 {
		player.InventoryLock.Lock()
		player.Inventory[player.HeldItem+36].ItemCount -= 1
		if player.Inventory[player.HeldItem+36].ItemCount == 0 {
			player.Inventory[player.HeldItem+36].Present = false
			player.Inventory[player.HeldItem+36].ItemID = 0
		}
		player.InventoryLock.Unlock()

		s.world.SetBlock(newX, newY, newZ, blockID)

		out := clientbound.BlockChange{
			Location: pk.Position{
				X: int32(newX),
				Y: int32(newY),
				Z: int32(newZ),
			},
			BlockID: pk.VarInt(blockID),
		}.CreatePacket()

		s.playerMap.Lock.RLock()
		for c := range s.playerMap.connToPlayer {
			// TODO: block change packet should only be sent to players if chunk is loaded
			_ = c.WritePacket(out)
		}
		s.playerMap.Lock.RUnlock()

		// send out updated item count
		player.InventoryLock.RLock()
		if err := conn.WritePacket(clientbound.SetSlot{
			WindowID: 0,
			Slot:     pk.Short(player.HeldItem + 36),
			SlotData: player.Inventory[player.HeldItem+36],
		}.CreatePacket()); err != nil {
			player.InventoryLock.RUnlock()
			return err
		}
		player.InventoryLock.RUnlock()
	}

	return nil
}
