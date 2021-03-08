package server

import (
	"github.com/panjf2000/gnet"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/serverbound"
	"math"
)

func (s *Server) handlePlayerBlockPlacement(_ gnet.Conn, pkt pk.Packet) (out []byte, err error) {
	in := serverbound.PlayerBlockPlacement{}
	if err = in.FromPacket(pkt); err != nil {
		return
	}

	location := in.Location
	newX := int(math.Floor(float64(location.X) + float64(in.CursorPositionX)))
	newY := int(math.Floor(float64(location.Y) + float64(in.CursorPositionY)))
	newZ := int(math.Floor(float64(location.Z) + float64(in.CursorPositionZ)))

	if in.CursorPositionX == 0 {
		newX -= 1
	}
	if in.CursorPositionY == 0 {
		newY -= 1
	}
	if in.CursorPositionZ == 0 {
		newZ -= 1
	}

	// TODO: determine block id from player inventory
	s.world.SetBlock(newX, newY, newZ, 1)

	// TODO: send block change packet to all players

	return
}
