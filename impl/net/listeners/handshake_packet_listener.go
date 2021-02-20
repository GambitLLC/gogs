package listeners

import (
	"errors"
	"github.com/panjf2000/gnet"
	pk "gogs/net/packet"
	"log"
)

type ConnectionState int8

const (
	status = 1
	login = 2
)

type HandshakePacketListener struct {
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
		c.SetContext(StatusPacketListener{int32(protocolVersion)})
	case login:
		c.SetContext(LoginPacketListener{protocolVersion: int32(protocolVersion)})
	default:
		log.Printf("Unhandled state %v", nextState)
		return errors.New("unhandled state")
	}

	return nil
}