package server

import (
	"fmt"
	"github.com/GambitLLC/gogs/net"
	pk "github.com/GambitLLC/gogs/net/packet"
	"github.com/GambitLLC/gogs/net/packet/clientbound"
)

func (s *Server) handleAnimation(conn net.Conn, pkt pk.Packet) (err error) {
	var hand pk.VarInt
	if err = pkt.Unmarshal(&hand); err != nil {
		return
	}

	if hand != 0 && hand != 1 {
		_ = conn.Close() // TODO: send disconnect packet
		err = fmt.Errorf("animation handler got hand %d", hand)
		return
	}

	player := s.playerFromConn(conn)

	anim := 0 // swing main arm
	if hand == 1 {
		anim = 3 // swing off hand
	}

	entityAnimationPacket := clientbound.EntityAnimation{
		EntityID:  pk.VarInt(player.ID()),
		Animation: pk.UByte(anim),
	}.CreatePacket()

	s.broadcastPacket(entityAnimationPacket, conn)

	return
}
