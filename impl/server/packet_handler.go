package server

import (
	"fmt"
	"github.com/panjf2000/gnet"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/packetids"
	"gogs/impl/net/packet/serverbound"
	"log"
)

type connectionState uint8

const (
	handshakeState connectionState = 0
	statusState                    = 1
	loginState                     = 2
	playState                      = 3
)

// Context for any given connection
type connectionContext struct {
	State           connectionState
	ProtocolVersion uint32
}

func (s *Server) handlePacket(conn gnet.Conn, pkt pk.Packet) ([]byte, error) {
	ctx := conn.Context().(connectionContext)
	switch ctx.State {
	case handshakeState:
		return s.handleHandshakeState(conn, pkt)
	case statusState:
		return s.handleStatusState(conn, pkt)
	case loginState:
		return s.handleLoginState(conn, pkt)
	case playState:
		return s.handlePlayState(conn, pkt)
	default:
		// shouldn't ever occur
		log.Panicf("invalid context state %d", ctx.State)
	}
	return nil, nil
}

func (s *Server) handleHandshakeState(conn gnet.Conn, pkt pk.Packet) (out []byte, err error) {
	if pkt.ID != 0 {
		return nil, fmt.Errorf("handshake expects Packet ID 0")
	}

	var (
		protocolVersion pk.VarInt
		address         pk.String
		port            pk.UShort
		nextState       pk.VarInt
	)

	err = pkt.Unmarshal(&protocolVersion, &address, &port, &nextState)
	if err != nil {
		return
	}

	logger.Printf("Received handshake: protocol %d and next state %d", protocolVersion, nextState)
	switch connectionState(nextState) {
	case statusState:
		conn.SetContext(connectionContext{
			State:           statusState,
			ProtocolVersion: uint32(protocolVersion),
		})
	case loginState:
		conn.SetContext(connectionContext{
			State:           loginState,
			ProtocolVersion: uint32(protocolVersion),
		})
	default:
		err = fmt.Errorf("handshake received invalid next state: %d", nextState)
	}

	return
}

func (s *Server) handleStatusState(conn gnet.Conn, pkt pk.Packet) (out []byte, err error) {
	switch pkt.ID {
	case packetids.StatusRequest:
		return s.handleStatusRequest()
	case packetids.StatusPing:
		logger.Printf("Received status ping packet")
		ping := serverbound.QueryStatusPing{}
		if err = ping.FromPacket(pkt); err != nil {
			return
		}

		out = clientbound.StatusPong{
			Payload: ping.Payload,
		}.CreatePacket().Encode()
	default:
		err = fmt.Errorf("status state received illegal packet id: 0x%02X", pkt.ID)
		_ = conn.Close()
	}
	return
}

func (s *Server) handleLoginState(conn gnet.Conn, pkt pk.Packet) (out []byte, err error) {

	switch pkt.ID {
	case packetids.LoginStart:
		return s.handleLoginStart(conn, pkt)
	case packetids.EncryptionResponse:
		err = fmt.Errorf("login state encryption not yet implemented")
	default:
		err = fmt.Errorf("login state received illegal packet id: 0x%02X", pkt.ID)
		_ = conn.Close()
	}
	return
}

func (s *Server) handlePlayState(conn gnet.Conn, pkt pk.Packet) (out []byte, err error) {
	switch pkt.ID {
	case packetids.TeleportConfirm:
		// TODO: Handle this
		logger.Printf("Received teleport confirm")
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
	case packetids.PlayerBlockPlacement:
		return s.handlePlayerBlockPlacement(conn, pkt)
	case packetids.HeldItemChangeServerbound:
		var slot pk.Short
		if err = pkt.Unmarshal(&slot); err != nil {
			return
		}
		player := s.playerFromConn(conn)
		player.HeldItem = uint8(slot)
		return
	case packetids.KeepAliveServerbound:
		logger.Printf("Received keep alive")
		//TODO: kick client for incorrect / untimely Keep-Alive response
		s := serverbound.KeepAlive{}
		if err := s.FromPacket(pkt); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("packet id 0x%02X not yet implemented", pkt.ID)
	}

	return
}
