package events

import (
	"github.com/panjf2000/gnet"
	"gogs/api/game"
)

type PlayerLoginResult int8

const (
	LoginAllowed PlayerLoginResult = iota
	KickBanned
	KickFull
	KickOther
	KickWhitelist
)

var PlayerLoginEvent playerLoginEvent

type PlayerLoginData struct {
	Player      *game.Player
	Conn        gnet.Conn
	Result      PlayerLoginResult
	KickMessage string
}

type playerLoginEvent struct {
	handlers   []func(*PlayerLoginData)
	netHandler func(*PlayerLoginData)
}

func (e *playerLoginEvent) Register(handler func(*PlayerLoginData)) {
	e.handlers = append([]func(*PlayerLoginData){handler}, e.handlers...)
}

func (e *playerLoginEvent) RegisterNet(handler func(*PlayerLoginData)) {
	e.netHandler = handler
}

func (e *playerLoginEvent) Trigger(data *PlayerLoginData) {
	for _, handler := range e.handlers {
		handler(data)
	}
	if e.netHandler != nil {
		e.netHandler(data)
	}
}
