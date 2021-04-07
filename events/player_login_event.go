package events

import (
	"gogs/net"
)

type PlayerLoginResult int8

const (
	LoginAllowed PlayerLoginResult = iota
	KickBanned
	KickFull
	KickOther
	KickWhitelist
)

// Triggered early in login process. Checks for whitelist/ban should be done here.
var PlayerLoginEvent playerLoginEvent

type PlayerLoginData struct {
	Name        string
	Conn        net.Conn
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

func (e *playerLoginEvent) Trigger(data *PlayerLoginData) {
	for _, handler := range e.handlers {
		handler(data)
	}
}
