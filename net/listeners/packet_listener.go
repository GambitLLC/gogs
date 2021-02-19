package listeners

import (
	"github.com/panjf2000/gnet"
	pk "gogs/net/packet"
)

type PacketListener interface {
	HandlePacket(gnet.Conn, *pk.Packet) error
}