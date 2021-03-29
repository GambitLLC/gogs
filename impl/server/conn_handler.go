package server

import (
	"fmt"
	"gogs/impl/logger"
	"gogs/impl/net"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/packetids"
	"gogs/impl/net/packet/serverbound"
)

type connectionState uint8

const (
	handshakeState connectionState = 0
	statusState                    = 1
	loginState                     = 2
	playState                      = 3
)

func (s *Server) handleHandshake(conn net.Conn) {
	pkt, err := conn.ReadPacket()
	if err != nil {
		logger.Printf("handshake error: %v", err)
		_ = conn.Close()
		return
	}

	if pkt.ID != 0 {
		logger.Printf("handshake expects Packet ID 0")
		_ = conn.Close()
		return
	}

	var (
		protocolVersion pk.VarInt
		address         pk.String
		port            pk.UShort
		nextState       pk.VarInt
	)

	err = pkt.Unmarshal(&protocolVersion, &address, &port, &nextState)
	if err != nil {
		logger.Printf("Handshake error: %v", err)
		_ = conn.Close()
		return
	}

	logger.Printf("Received handshake: protocol %d and next state %d", protocolVersion, nextState)
	switch connectionState(nextState) {
	case statusState:
		err = s.handleStatus(conn)
		if err != nil {
			logger.Printf("status state error: %v", err)
		}
	case loginState:
		err = s.handleLogin(conn)
		if err != nil {
			logger.Printf("login state received error: %v", err)
			break
		}
		err = s.handlePlay(conn)
		if err != nil {
			logger.Printf("play state received error: %v", err)
		}
	default:
		logger.Printf("handshake received invalid next state: %d", nextState)
	}

	// close connection after handlers are done ... note handlePlay is blocking so this goroutine won't end early
	logger.Printf("closing??? err: %v", err)
	_ = conn.Close()
	// TODO: handle closed connection stuff
	s.handleClosedState(conn)
}

func (s *Server) handleClosedState(conn net.Conn) {
	logger.Printf("Connection closed")

	//clean up all the player state
	s.playerMap.Lock.RLock()
	player, exists := s.playerMap.connToPlayer[conn]
	s.playerMap.Lock.RUnlock()

	if exists {
		s.playerMap.Lock.Lock()
		delete(s.playerMap.connToPlayer, conn)
		s.playerMap.Lock.Unlock()

		player.Connection = nil
		player.Online = false

		logger.Printf("Player %v disconnected", player.Name)

		// update player info for all remaining players
		playerInfoPacket := clientbound.PlayerInfo{
			Action:     4, // TODO: create consts for action
			NumPlayers: 1,
			Players: []pk.Encodable{
				clientbound.PlayerInfoRemovePlayer{
					UUID: pk.UUID(player.UUID),
				},
			},
		}.CreatePacket()
		// also destroy the entity for all players
		destroyEntitiesPacket := clientbound.DestroyEntities{
			Count:     1,
			EntityIDs: []pk.VarInt{pk.VarInt(player.ID())},
		}.CreatePacket()

		s.playerMap.Lock.RLock()
		for c := range s.playerMap.connToPlayer {
			_ = c.WritePacket(playerInfoPacket)
			_ = c.WritePacket(destroyEntitiesPacket)
		}
		s.playerMap.Lock.RUnlock()

		// TODO: trigger disconnect event
		s.Broadcast(fmt.Sprintf("%v has left the game", player.Name))
	}
}

func (s *Server) handleStatus(conn net.Conn) error {
	pkt, err := conn.ReadPacket()
	if err != nil {
		return err
	}

	switch pkt.ID {
	case packetids.StatusRequest:
		resp, err := s.statusPacket()
		if err != nil {
			return err
		}
		return conn.WritePacket(resp)
	case packetids.StatusPing:
		logger.Printf("Received status ping packet")
		ping := serverbound.QueryStatusPing{}
		if err = ping.FromPacket(pkt); err != nil {
			return err
		}

		return conn.WritePacket(clientbound.StatusPong{
			Payload: ping.Payload,
		}.CreatePacket())

	default:
		return fmt.Errorf("status state received illegal packet id: 0x%02X", pkt.ID)
	}
}

func (s *Server) handleLogin(conn net.Conn) error {
	pkt, err := conn.ReadPacket()
	if err != nil {
		return err
	}

	if pkt.ID != packetids.LoginStart {
		return fmt.Errorf("login state expected LoginStart, received 0x%02X instead", pkt.ID)
	}

	return s.handleLoginStart(conn, pkt)
}

func (s *Server) handlePlay(conn net.Conn) (err error) {
	// block this goroutine to keep connection up
	var pkt pk.Packet
	for {
		pkt, err = conn.ReadPacket()
		if err != nil {
			return err
		}

		switch pkt.ID {
		case packetids.ClientSettings:
			err = s.handleClientSettings(conn, pkt)
		case packetids.TeleportConfirm:
			// TODO: Handle this
			logger.Printf("Received teleport confirm")
			/*
				case packetids.ChatMessageServerbound:
					return s.handleChatMessage(conn, pkt)
				case packetids.ClientSettings:
					return s.handleClientSettings(conn, pkt)
				case packetids.PlayerPosition:
					// TODO: Handle all player pos & rotation packets
					return s.handlePlayerPosition(conn, pkt)
				case packetids.PlayerPositionAndRotationServerbound:
					return s.handlePlayerPositionAndRotation(conn, pkt)
				case packetids.PlayerRotation:
					return s.handlePlayerRotation(conn, pkt)
				case packetids.Animation:
					return s.handleAnimation(conn, pkt)
				case packetids.EntityAction:
					return s.handleEntityAction(conn, pkt)
				case packetids.InteractEntity:
					return s.handleInteractEntity(conn, pkt)
				case packetids.ClientStatus:
					return s.handleClientStatus(conn, pkt)
				case packetids.ClickWindow:
					return s.handleClickWindow(conn, pkt)
				case packetids.PlayerDigging:
					return s.handlePlayerDigging(conn, pkt)
				case packetids.PlayerBlockPlacement:
					return s.handlePlayerBlockPlacement(conn, pkt)

			*/
		case packetids.HeldItemChangeServerbound:
			var slot pk.Short
			if err = pkt.Unmarshal(&slot); err != nil {
				return err
			}
			player := s.playerFromConn(conn)
			player.HeldItem = uint8(slot)

		case packetids.KeepAliveServerbound:
			logger.Printf("Received keep alive")
			//TODO: kick client for incorrect / untimely Keep-Alive response
			k := serverbound.KeepAlive{}
			if err = k.FromPacket(pkt); err != nil {
				return
			}
		default:
			logger.Printf("packet id 0x%02X not yet implemented", pkt.ID)
		}

		if err != nil {
			return err
		}
	}
}
