package serverbound

import pk "gogs/impl/net/packet"

type EntityAction struct {
	EntityID  pk.VarInt
	ActionID  pk.VarInt
	JumpBoost pk.VarInt
}

func (p *EntityAction) FromPacket(packet pk.Packet) error {
	return packet.Unmarshal(&p.EntityID, &p.ActionID, &p.JumpBoost)
}
