package handlers

import (
	"fmt"
	"github.com/panjf2000/gnet"
	"gogs/api"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/packetids"
)

func Animation(c gnet.Conn, pkt *pk.Packet, s api.Server) error {
	if pkt.ID != packetids.Animation {
		return fmt.Errorf("animation handler expects packet id %x, got %x", packetids.Animation, pkt.ID)
	}
	var hand pk.VarInt
	if err := pkt.Unmarshal(&hand); err != nil {
		return err
	}

	if hand != 0 && hand != 1 {
		_ = c.Close() // TODO: send disconnect packet
		return fmt.Errorf("animation handler got hand %d", hand)
	}

	player := s.PlayerFromConn(c)

	anim := 0 // swing main arm
	if hand == 1 {
		anim = 3 // swing off hand
	}

	entityAnimationPacket := clientbound.EntityAnimation{
		EntityID:  pk.VarInt(player.EntityID),
		Animation: pk.UByte(anim),
	}.CreatePacket().Encode()

	for _, p := range s.Players() {
		conn := s.ConnFromUUID(p.UUID)
		if conn != c {
			conn.AsyncWrite(entityAnimationPacket)
		}
	}

	return nil
}
