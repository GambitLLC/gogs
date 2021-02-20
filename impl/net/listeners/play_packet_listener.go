package listeners

import (
	"errors"
	"github.com/panjf2000/gnet"
	"gogs/api"
	pk "gogs/impl/net/packet"
)

type PlayPacketListener struct {
	S api.Server
	protocolVersion int32
}

func (listener PlayPacketListener) HandlePacket(c gnet.Conn, p *pk.Packet) error {
	return errors.New("not yet implemented")
}