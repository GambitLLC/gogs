package api

import (
	"bedgg-server/api/events"
)

type EventHandler struct {
}

func (eh EventHandler) handleEvent(event events.Event) {
	event.Handle()
}
