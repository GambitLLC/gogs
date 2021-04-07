package serverbound

import pk "gogs/net/packet"

type QueryStatusPing struct {
	Payload pk.Long
}

func (p *QueryStatusPing) FromPacket(packet pk.Packet) error {
	return packet.Unmarshal(&p.Payload)
}
