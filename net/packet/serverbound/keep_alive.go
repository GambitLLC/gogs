package serverbound

import pk "github.com/GambitLLC/gogs/net/packet"

type KeepAlive struct {
	ID pk.Long
}

func (p *KeepAlive) FromPacket(packet pk.Packet) error {
	return packet.Unmarshal(&p.ID)
}
