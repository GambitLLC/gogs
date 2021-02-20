package listeners

import (
	"errors"
	"github.com/panjf2000/gnet"
	pk "gogs/net/packet"
)

type StatusPacketListener struct {
	protocolVersion int32
}

func (listener StatusPacketListener) HandlePacket(c gnet.Conn, p *pk.Packet) error {
	return errors.New("not yet implemented")
}