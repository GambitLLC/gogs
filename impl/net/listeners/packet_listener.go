package listeners

import (
	"github.com/panjf2000/gnet"
	pk "gogs/impl/net/packet"
)

type PacketListener interface {
	HandlePacket(gnet.Conn, *pk.Packet) ([]byte, error)
}
