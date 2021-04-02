package server

import (
	"fmt"
	"gogs/api/data/chat"
	"gogs/impl/ecs"
	"gogs/impl/logger"
	"gogs/impl/net"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/serverbound"
)

func (s *Server) handleInteractEntity(conn net.Conn, pkt pk.Packet) (err error) {
	in := serverbound.InteractEntity{}
	if err = in.FromPacket(pkt); err != nil {
		return
	}

	logger.Printf("received interact entity")

	switch in.Type {
	case 0: // interact
	case 1: // attack
		player := s.entityFromID(uint64(in.EntityID)).(*ecs.Player)
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

func (s *Server) handleAttack(attacker *ecs.Player, defender *ecs.Player) {
	// TODO: check if attacker is in range of defender
	// TODO: handle death stuff
	defender.Health -= 2
	remainingHealth := defender.Health
	outPacket := clientbound.UpdateHealth{
		Health:         pk.Float(remainingHealth),
		Food:           20.0,
		FoodSaturation: 0,
	}.CreatePacket()

	_ = defender.Connection.WritePacket(outPacket)

	if remainingHealth == 0 {
		out := clientbound.EntityStatus{
			EntityID:     pk.Int(defender.ID()),
			EntityStatus: 3, // LivingEntity play death sound and animation
		}.CreatePacket()
		s.broadcastPacket(out, nil)

		_ = defender.Connection.WritePacket(clientbound.CombatEvent{
			PlayerID: pk.VarInt(defender.ID()),
			EntityID: pk.Int(attacker.ID()),
			Message:  pk.Chat(chat.NewMessage("You have died").AsJSON()),
		}.CreatePacket())
		return
	}

	// TODO: figure out what sounds are played and for who
	_ = attacker.Connection.WritePacket(clientbound.NamedSoundEffect{
		SoundName:       "entity.player.attack.strong",
		SoundCategory:   7, // Player category?
		EffectPositionX: pk.Int(defender.X * 8),
		EffectPositionY: pk.Int(defender.Y * 8),
		EffectPositionZ: pk.Int(defender.Z * 8),
		Volume:          1.0,
		Pitch:           1.0,
	}.CreatePacket())

	// TODO: this should only broadcast to players in visual range
	out := clientbound.EntityStatus{
		EntityID:     pk.Int(defender.ID()),
		EntityStatus: 2, // LivingEntity play hurt sound and animation
	}.CreatePacket()
	s.broadcastPacket(out, nil)

}
