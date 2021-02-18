package events

import (
	"bedgg-server/api/game"
	emitter "github.com/emitter-io/go/v2"
)

/* Player Join Event */
type IPlayerJoinEvent interface {
	Event
	getPlayer() game.Player
}

type PlayerJoinEvent struct {
	IPlayerJoinEvent
	player    game.Player
	emitter   *emitter.Client
	eventType EventType
}

func (event PlayerJoinEvent) getPlayer() game.Player {
	return event.player
}

func (event PlayerJoinEvent) getEmitter() *emitter.Client {
	return event.emitter
}

func (event PlayerJoinEvent) getEventType() EventType {
	return event.eventType
}

func (event PlayerJoinEvent) handle() {
	//TODO: Handle Player join event
}
