package events

import (
	"github.com/GambitLLC/gogs/chat"
	"github.com/GambitLLC/gogs/entities"
)

var PlayerChatEvent playerChatEvent

type PlayerChatData struct {
	Player     entities.Player
	Recipients []entities.Player
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

func (e *playerChatEvent) Trigger(data *PlayerChatData) {
	for _, handler := range e.handlers {
		handler(data)
	}
}
