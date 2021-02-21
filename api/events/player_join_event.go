package events

import "gogs/api/game"

var PlayerJoinEvent playerJoinEvent

type PlayerJoinData struct {
	Player  *game.Player
	Message string
}

type playerJoinEvent struct {
	handlers   []func(*PlayerJoinData)
	netHandler func(*PlayerJoinData)
}

func (e *playerJoinEvent) Register(handler func(*PlayerJoinData)) {
	e.handlers = append([]func(*PlayerJoinData){handler}, e.handlers...)
}

func (e *playerJoinEvent) RegisterNet(handler func(*PlayerJoinData)) {
	e.netHandler = handler
}

func (e *playerJoinEvent) Trigger(data *PlayerJoinData) {
	for _, handler := range e.handlers {
		handler(data)
	}
	if e.netHandler != nil {
		e.netHandler(data)
	}
}
