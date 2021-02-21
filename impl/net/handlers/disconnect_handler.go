package handlers

import (
	"fmt"
	"gogs/api"
	"gogs/api/game"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
)

// Called when a connection is closed
func Disconnect(player game.Player, s api.Server) error {
	logger.Printf("Player %v disconnected", player.Name)

	// update player info for all remaining players
	playerInfoPacket := clientbound.PlayerInfo{
		Action:     4, // TODO: create consts for action
		NumPlayers: 1,
		Players: []pk.Encodable{
			clientbound.PlayerInfoRemovePlayer{
				UUID: pk.UUID(player.UUID),
			},
		},
	}.CreatePacket().Encode()
	// also destroy the entity for all players
	destroyEntitiesPacket := clientbound.DestroyEntities{
		Count:     1,
		EntityIDs: []pk.VarInt{pk.VarInt(player.EntityID)},
	}.CreatePacket().Encode()
	for _, p := range s.Players() {
		_ = s.ConnFromUUID(p.UUID).AsyncWrite(append(playerInfoPacket, destroyEntitiesPacket...))
	}

	// TODO: trigger disconnect event
	s.Broadcast(fmt.Sprintf("%v has left the game", player.Name))

	return nil
}
