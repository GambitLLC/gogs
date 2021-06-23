package server

import (
	"github.com/GambitLLC/gogs/chat"
	"github.com/GambitLLC/gogs/entities"
	pk "github.com/GambitLLC/gogs/net/packet"
	"github.com/GambitLLC/gogs/net/packet/clientbound"
)

// TODO: move this into player?
func (s *Server) handleAttack(attacker *entities.Player, defender *entities.Player) (err error) {
	// TODO: check if attacker is in range of defender
	if defender.Health == 0 {
		return
	}

	defender.Health -= 2
	remainingHealth := defender.Health

	if err = defender.Connection.WritePacket(clientbound.UpdateHealth{
		Health:         pk.Float(remainingHealth),
		Food:           pk.VarInt(defender.Food),
		FoodSaturation: pk.Float(defender.Saturation),
	}.CreatePacket()); err != nil {
		return
	}

	if remainingHealth == 0 {
		// updating health tells the client to play death animation
		s.broadcastPacket(clientbound.EntityMetadata{
			EntityID: pk.VarInt(defender.ID()),
			Metadata: []clientbound.MetadataField{
				{Index: 8, Type: 2, Value: pk.Float(defender.Health)}, // HEALTH
				{Index: 0xFF},
			},
		}.CreatePacket(), defender.Connection)

		err = defender.Connection.WritePacket(clientbound.CombatEvent{
			PlayerID: pk.VarInt(defender.ID()),
			EntityID: pk.Int(attacker.ID()),
			Message:  pk.Chat(chat.NewStringComponent("You have died").AsJSON()),
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

	return
}
