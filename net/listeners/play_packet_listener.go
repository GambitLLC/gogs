package listeners

import (
	"errors"
	"github.com/panjf2000/gnet"
	pk "gogs/net/packet"
)

type playPacketListener struct {
	protocolVersion int32
}

func PlayPacketListener(protoVersion int32) playPacketListener {
	return playPacketListener{protoVersion}
}

func (listener playPacketListener) HandlePacket(c gnet.Conn, p *pk.Packet) error {
	return errors.New("not yet implemented")
}