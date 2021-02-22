package listeners

import (
	"errors"
	"github.com/panjf2000/gnet"
	"gogs/api"
	"gogs/impl/logger"
	"gogs/impl/net/handlers"
	pk "gogs/impl/net/packet"
)

type LoginState int8

const (
	start LoginState = iota
	encrypt
)

type LoginPacketListener struct {
	S               api.Server
	protocolVersion int32
	encrypt         bool
	state           LoginState
}

func (listener LoginPacketListener) HandlePacket(c gnet.Conn, p *pk.Packet) (out []byte, err error) {
	switch listener.state {
	case start:
		if out, err = handlers.LoginStart(c, p, listener.S); err != nil {
			_ = c.Close()	// LoginStart errors won't return any output
			return
		}
		c.SetContext(PlayPacketListener{
			S:               listener.S,
			protocolVersion: listener.protocolVersion,
		})
	case encrypt:
		// TODO: implement encryption
		return nil, errors.New("login encryption is not yet implemented")
	default:
		logger.Printf("LoginPacketListener is in an unknown state: %d", listener.state)
		return nil, c.Close()
	}
	return
}
