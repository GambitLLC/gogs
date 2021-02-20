package listeners

import (
	"errors"
	"github.com/panjf2000/gnet"
	"gogs/api"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
)

type ConnectionState int8

const (
	status = 1
	login  = 2
)

type HandshakePacketListener struct {
	S api.Server
}

func (listener HandshakePacketListener) HandlePacket(c gnet.Conn, p *pk.Packet) error {
	if p.ID != 0 {
		return errors.New("handshake expects Packet ID 0")
	}

	var (
		protocolVersion pk.VarInt
		address         pk.String
		port            pk.UShort
		nextState       pk.VarInt
	)

	err := p.Unmarshal(&protocolVersion, &address, &port, &nextState)
	if err != nil {
		return err
	}

	switch ConnectionState(nextState) {
	case status:
		c.SetContext(StatusPacketListener{
			S:               listener.S,
			protocolVersion: int32(protocolVersion),
		})
	case login:
		c.SetContext(LoginPacketListener{
			S:               listener.S,
			protocolVersion: int32(protocolVersion),
		})
	default:
		logger.Printf("Unhandled state %v", nextState)
		return errors.New("unhandled state")
	}

	return nil
}
