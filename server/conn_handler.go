package server

import (
	"fmt"
	"github.com/GambitLLC/gogs/chat"
	"github.com/GambitLLC/gogs/logger"
	"github.com/GambitLLC/gogs/net"
	pk "github.com/GambitLLC/gogs/net/packet"
	"github.com/GambitLLC/gogs/net/packet/clientbound"
	"github.com/GambitLLC/gogs/net/packet/packetids"
	"github.com/GambitLLC/gogs/net/packet/serverbound"
	"io"
	"time"
)

type connectionState uint8

const (
	handshakeState connectionState = 0
	statusState                    = 1
	loginState                     = 2
	playState                      = 3
)

func (c connectionState) String() string {
	switch c {
	case handshakeState:
		return "handshake state"
	case statusState:
		return "status state"
	case loginState:
		return "login state"
	case playState:
		return "play state"
	default:
		return "invalid connection state"
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	state := handshakeState
	var err error

	state, err = s.handleHandshake(conn)
	if err != nil {
		state = 0
		goto close
	}
	switch state {
	case statusState:
		err = s.handleStatus(conn)
		if err != nil {
			goto close
		}
	case loginState:
		err = s.handleLogin(conn)
		if err != nil {
			goto close
		}

		state = playState
		err = s.handlePlay(conn)
		goto close
	}

close:
	select {
	case <-s.shutdown:
		return
	default:
		// ignore EOF (connection closed by client)
		if err != nil && err != io.EOF {
			logger.Printf("%s error: %v", state.String(), err)
		}

		_ = conn.Close()
		s.removeConnection(conn)
	}
}

func (s *Server) handleHandshake(conn net.Conn) (connectionState, error) {
	pkt, err := conn.ReadPacket()
	if err != nil {
		return 0, err
	}

	if pkt.ID != packetids.Handshake {
		return 0, fmt.Errorf("expected handshake, got packetid 0x%x", pkt.ID)
	}

	var (
		protocolVersion pk.VarInt
		address         pk.String
		port            pk.UShort
		nextState       pk.VarInt
	)

	err = pkt.Unmarshal(&protocolVersion, &address, &port, &nextState)
	if err != nil {
		return 0, err
	}

	if nextState != statusState && nextState != loginState {
		return 0, fmt.Errorf("received invalid next state %d", nextState)
	}

	logger.Printf("Received handshake: protocol %d and next state %d", protocolVersion, nextState)

	return connectionState(nextState), nil
}

// removeConnection clears details related to the player for this connection. Does not close the connection.
func (s *Server) removeConnection(conn net.Conn) {
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

func (s *Server) handleStatus(conn net.Conn) (err error) {
	var pkt pk.Packet
	pkt, err = conn.ReadPacket()
	if err != nil {
		return err
	}

	if pkt.ID != packetids.StatusRequest {
		return fmt.Errorf("status state expected StatusRequest, got %d", pkt.ID)
	}

	// send status response
	if pkt, err = s.statusPacket(); err != nil {
		return
	}
	if err = conn.WritePacket(pkt); err != nil {
		return
	}

	pkt, err = conn.ReadPacket()
	if err != nil {
		return err
	}

	if pkt.ID != packetids.StatusPing {
		return fmt.Errorf("status state expected StatusPing, got %d", pkt.ID)
	}

	ping := serverbound.QueryStatusPing{}
	if err = ping.FromPacket(pkt); err != nil {
		return err
	}

	// send pong
	return conn.WritePacket(clientbound.StatusPong{
		Payload: ping.Payload,
	}.CreatePacket())
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
	var pkt pk.Packet
	var keepAliveID, receivedKeepAliveID int64
	var alive = true

	// send keep alive packets
	go func() {
		var t time.Time
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for ; true; t = <-ticker.C {
			// didn't receive a timely keep alive
			if receivedKeepAliveID != keepAliveID {
				alive = false
				_ = conn.WritePacket(pk.Marshal(
					packetids.PlayDisconnect,
					pk.Chat(chat.NewStringComponent("Kicked due to keep alive timeout").AsJSON()),
				))
				_ = conn.Close()
			}

			keepAliveID = t.UnixNano()
			if err = conn.WritePacket(clientbound.KeepAlive{
				ID: pk.Long(keepAliveID),
			}.CreatePacket()); err != nil {
				_ = conn.Close()
				return
			}
		}
	}()

	// block this goroutine to keep connection up
	for {
		pkt, err = conn.ReadPacket()
		if err != nil {
			if !alive {
				err = nil // connection was closed due to keep alive timeout
			}
			return
		}

		switch pkt.ID {
		case packetids.ClientSettings:
			err = s.handleClientSettings(conn, pkt)
		case packetids.TeleportConfirm:
			// TODO: Handle this
			logger.Printf("Received teleport confirm")
		case packetids.ChatMessageServerbound:
			err = s.handleChatMessage(conn, pkt)
		case packetids.PlayerPosition:
			// TODO: Handle all player pos & rotation packets
			err = s.handlePlayerPosition(conn, pkt)
		case packetids.PlayerPositionAndRotationServerbound:
			err = s.handlePlayerPositionAndRotation(conn, pkt)
		case packetids.PlayerRotation:
			err = s.handlePlayerRotation(conn, pkt)
		case packetids.Animation:
			err = s.handleAnimation(conn, pkt)
		case packetids.EntityAction:
			err = s.handleEntityAction(conn, pkt)
		case packetids.InteractEntity:
			err = s.handleInteractEntity(conn, pkt)
		case packetids.ClientStatus:
			err = s.handleClientStatus(conn, pkt)
		case packetids.ClickWindow:
			err = s.handleClickWindow(conn, pkt)
		case packetids.PlayerDigging:
			err = s.handlePlayerDigging(conn, pkt)
		case packetids.PlayerBlockPlacement:
			err = s.handlePlayerBlockPlacement(conn, pkt)
		case packetids.HeldItemChangeServerbound:
			var slot pk.Short
			if err = pkt.Unmarshal(&slot); err != nil {
				return
			}
			player := s.playerFromConn(conn)
			player.HeldItem = uint8(slot)
		case packetids.KeepAliveServerbound:
			logger.Printf("Received keep alive")
			k := serverbound.KeepAlive{}
			if err = k.FromPacket(pkt); err != nil {
				return
			}

			if receivedKeepAliveID = int64(k.ID); receivedKeepAliveID != keepAliveID {
				_ = conn.WritePacket(pk.Marshal(
					packetids.PlayDisconnect,
					pk.Chat(chat.NewStringComponent("Kicked due to invalid keep alive ID").AsJSON()),
				))
				_ = conn.Close()
				return nil
			}
		default:
			logger.Printf("packet id 0x%02X not yet implemented", pkt.ID)
		}

		if err != nil {
			return
		}
	}
}
