package events

import (
	"gogs/entities"
)

var PlayerJoinEvent playerJoinEvent

type PlayerJoinData struct {
	Player  entities.Player
	Message string
}

type playerJoinEvent struct {
	handlers   []func(*PlayerJoinData)
	netHandler func(*PlayerJoinData)
}

func (e *playerJoinEvent) Register(handler func(*PlayerJoinData)) {
	e.handlers = append([]func(*PlayerJoinData){handler}, e.handlers...)
}

func (e *playerJoinEvent) Trigger(data *PlayerJoinData) {
	for _, handler := range e.handlers {
		handler(data)
	}
}
