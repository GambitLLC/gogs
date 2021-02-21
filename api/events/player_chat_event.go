package events

import (
	"gogs/api/data/chat"
	"gogs/api/game"
)

var PlayerChatEvent playerChatEvent

type PlayerChatData struct {
	Player     *game.Player
	Recipients []*game.Player
	Message    chat.Message
	Format     string // First argument will be Player.Name, second will be Message
}

type playerChatEvent struct {
	handlers   []func(*PlayerChatData)
	netHandler func(*PlayerChatData)
}

func (e *playerChatEvent) Register(handler func(*PlayerChatData)) {
	e.handlers = append([]func(*PlayerChatData){handler}, e.handlers...)
}

func (e *playerChatEvent) RegisterNet(handler func(*PlayerChatData)) {
	e.netHandler = handler
}

func (e *playerChatEvent) Trigger(data *PlayerChatData) {
	for _, handler := range e.handlers {
		handler(data)
	}
	if e.netHandler != nil {
		e.netHandler(data)
	}
}
