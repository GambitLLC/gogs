package server

import (
	"fmt"
	"github.com/GambitLLC/gogs/chat"
	"github.com/GambitLLC/gogs/entities"
	"github.com/GambitLLC/gogs/logger"
	"github.com/GambitLLC/gogs/net"
	pk "github.com/GambitLLC/gogs/net/packet"
	"github.com/GambitLLC/gogs/net/packet/clientbound"
	"github.com/GambitLLC/gogs/net/packet/serverbound"
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
		player := s.entityFromID(uint64(in.EntityID)).(*entities.Player)
		if player == nil {
			err = fmt.Errorf("interact entity could not find entity with id %d", in.EntityID)
			break
		}
		err = s.handleAttack(s.playerFromConn(conn), player)
	case 2: // interact at
	default:
		_ = conn.Close()
		err = fmt.Errorf("interact entity got invalid type %d", in.Type)
	}
	return
}

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
