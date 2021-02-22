package handlers

import (
	"fmt"
	"github.com/panjf2000/gnet"
	"gogs/api"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
)

func Animation(c gnet.Conn, pkt *pk.Packet, s api.Server) ([]byte, error) {
	var hand pk.VarInt
	if err := pkt.Unmarshal(&hand); err != nil {
		return nil, err
	}

	if hand != 0 && hand != 1 {
		_ = c.Close() // TODO: send disconnect packet
		return nil, fmt.Errorf("animation handler got hand %d", hand)
	}

	player := s.PlayerFromConn(c)

	anim := 0 // swing main arm
	if hand == 1 {
		anim = 3 // swing off hand
	}

	entityAnimationPacket := clientbound.EntityAnimation{
		EntityID:  pk.VarInt(player.EntityID()),
		Animation: pk.UByte(anim),
	}.CreatePacket().Encode()

	// Send animation to everyone except self
	for _, p := range s.Players() {
		conn := s.ConnFromUUID(p.UUID())
		if conn != c {
			_ = conn.AsyncWrite(entityAnimationPacket)
		}
	}

	return nil, nil
}
