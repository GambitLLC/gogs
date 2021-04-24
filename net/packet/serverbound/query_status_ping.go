package serverbound

import pk "github.com/GambitLLC/gogs/net/packet"

type QueryStatusPing struct {
	Payload pk.Long
}

func (p *QueryStatusPing) FromPacket(packet pk.Packet) error {
	return packet.Unmarshal(&p.Payload)
}
