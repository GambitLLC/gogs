package api

import (
	"gogs/api/events"
)

type EventHandler struct {
}

func (eh EventHandler) handleEvent(event events.Event) {
	event.Handle()
}
