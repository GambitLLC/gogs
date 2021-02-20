package listeners

import (
	"errors"
	"github.com/panjf2000/gnet"
	pk "gogs/net/packet"
)

type PlayPacketListener struct {
	protocolVersion int32
}

func (listener PlayPacketListener) HandlePacket(c gnet.Conn, p *pk.Packet) error {
	return errors.New("not yet implemented")
}