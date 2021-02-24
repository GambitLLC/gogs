package api

import (
	"github.com/google/uuid"
	"gogs/api/game"
)

type Server interface {
	Players() []game.Player
	PlayerFromUUID(uuid.UUID) game.Player
	Broadcast(string)
}
