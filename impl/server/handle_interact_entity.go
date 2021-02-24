package server

import (
	"fmt"
	"github.com/panjf2000/gnet"
	"gogs/api/data/chat"
	"gogs/impl/game"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/serverbound"
)

func (s *Server) handleInteractEntity(conn gnet.Conn, pkt pk.Packet) (out []byte, err error) {
	in := serverbound.InteractEntity{}
	if err = in.FromPacket(pkt); err != nil {
		return
	}

	logger.Printf("received interact entity")

	switch in.Type {
	case 0: // interact
	case 1: // attack
		player := s.playerFromEntityID(int32(in.EntityID))
		if player == nil {
			err = fmt.Errorf("interact entity could not find entity with id %d", in.EntityID)
			break
		}
		s.handleAttack(s.playerFromConn(conn), player)
	case 2: // interact at
	default:
		_ = conn.Close()
		err = fmt.Errorf("interact entity got invalid type %d", in.Type)
	}
	return
}

func (s *Server) handleAttack(attacker *game.Player, defender *game.Player) {
	// TODO: check if attacker is in range of defender
	// TODO: handle death stuff
	*defender.Health() -= 2
	remainingHealth := *defender.Health()
	outPacket := clientbound.UpdateHealth{
		Health:         pk.Float(remainingHealth),
		Food:           20.0,
		FoodSaturation: 0,
	}.CreatePacket().Encode()

	_ = defender.Conn().AsyncWrite(outPacket)

	if remainingHealth == 0 {
		out := clientbound.EntityStatus{
			EntityID:     pk.Int(defender.EntityID()),
			EntityStatus: 3, // LivingEntity play death sound and animation
		}.CreatePacket()
		s.broadcastPacket(out, nil)

		_ = defender.Conn().AsyncWrite(clientbound.CombatEvent{
			PlayerID: pk.VarInt(defender.EntityID()),
			EntityID: pk.Int(attacker.EntityID()),
			Message:  pk.Chat(chat.NewMessage("You have died").AsJSON()),
		}.CreatePacket().Encode())
		return
	}

	// TODO: figure out what sounds are played and for who
	_ = attacker.Conn().AsyncWrite(clientbound.NamedSoundEffect{
		SoundName:       "entity.player.attack.strong",
		SoundCategory:   7, // Player category?
		EffectPositionX: pk.Int(defender.Position().X * 8),
		EffectPositionY: pk.Int(defender.Position().Y * 8),
		EffectPositionZ: pk.Int(defender.Position().Z * 8),
		Volume:          1.0,
		Pitch:           1.0,
	}.CreatePacket().Encode())

	// TODO: this should only broadcast to players in visual range
	out := clientbound.EntityStatus{
		EntityID:     pk.Int(defender.EntityID()),
		EntityStatus: 2, // LivingEntity play hurt sound and animation
	}.CreatePacket()
	s.broadcastPacket(out, nil)

}
