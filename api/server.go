package api

import (
	"github.com/google/uuid"
	"github.com/panjf2000/gnet"
	"gogs/api/game"
)

type Server interface {
	Players() []game.Player
	CreatePlayer(string, uuid.UUID, gnet.Conn) game.Player
	PlayerFromConn(gnet.Conn) game.Player
	PlayerFromUUID(uuid.UUID) game.Player
	ConnFromUUID(uuid.UUID) gnet.Conn
	Broadcast(string)
}
