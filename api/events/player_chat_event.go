package events

import (
	"encoding/json"
	"fmt"
	"gogs/api/game"
)

var PlayerChatEvent playerChatEvent

type PlayerChatData struct {
	Player     *game.Player
	Recipients []*game.Player
	Message    string
	Format     string // First argument will be Player.Name, second will be Message
}

type chatJSON struct {
	Message string      `json:"text"`
	Extra   []*chatJSON `json:"extra,omitempty"`
}

func (d PlayerChatData) AsJSON() string {
	if d.Format == "" {
		d.Format = "%s: %s"
	}
	chat := chatJSON{
		Message: fmt.Sprintf(d.Format, d.Player.Name, d.Message),
		Extra:   nil,
	}
	if text, err := json.Marshal(chat); err != nil {
		panic(err)
	} else {
		return string(text)
	}
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
