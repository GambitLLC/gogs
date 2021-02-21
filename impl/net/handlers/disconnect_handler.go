package handlers

import (
	"fmt"
	"github.com/panjf2000/gnet"
	"gogs/api"
	"gogs/api/game"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
)

// Called when a connection is closed
func Disconnect(c gnet.Conn, player game.Player, s api.Server) error {
	logger.Printf("Player %v disconnected", player)

	// update player info for all remaining players
	packet := clientbound.PlayerInfo{
		Action:     4, // TODO: create consts for action
		NumPlayers: 1,
		Players: []pk.Encodable{
			clientbound.PlayerInfoRemovePlayer{
				UUID: pk.UUID(player.UUID),
			},
		},
	}.CreatePacket().Encode()
	for _, p := range s.Players() {
		_ = s.ConnFromUUID(p.UUID).AsyncWrite(packet)
	}

	// TODO: trigger disconnect event
	s.Broadcast(fmt.Sprintf("%v has left the game", player.Name))

	return nil
}
