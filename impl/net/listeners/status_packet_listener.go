package listeners

import (
	"errors"
	"github.com/panjf2000/gnet"
	"gogs/impl"
	pk "gogs/net/packet"
)

type StatusPacketListener struct {
	S *impl.Server
	protocolVersion int32
}

func (listener StatusPacketListener) HandlePacket(c gnet.Conn, p *pk.Packet) error {
	return errors.New("not yet implemented")
}