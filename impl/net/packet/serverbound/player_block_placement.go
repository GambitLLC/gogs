package serverbound

import pk "gogs/impl/net/packet"

type PlayerBlockPlacement struct {
	Hand            pk.VarInt
	Location        pk.Position
	Face            pk.VarInt
	CursorPositionX pk.Float
	CursorPositionY pk.Float
	CursorPositionZ pk.Float
	InsideBlock     pk.Boolean
}

func (s *PlayerBlockPlacement) FromPacket(packet pk.Packet) error {
	return packet.Unmarshal(
		&s.Hand,
		&s.Location,
		&s.Face,
		&s.CursorPositionX,
		&s.CursorPositionY,
		&s.CursorPositionZ,
		&s.InsideBlock,
	)
}
