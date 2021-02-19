package listeners

import (
	"errors"
	"github.com/panjf2000/gnet"
	pk "gogs/net/packet"
)

type statusPacketListener struct {
	protocolVersion int32
}

func StatusPacketListener(protoVersion int32) statusPacketListener {
	return statusPacketListener{protoVersion}
}

func (listener statusPacketListener) HandlePacket(c gnet.Conn, p *pk.Packet) error {
	return errors.New("not yet implemented")
}