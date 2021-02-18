package events

import emitter "github.com/emitter-io/go/v2"

type EventType int

const (
	PLAYER_JOIN_EVENT EventType = 0
	PLAYER_QUIT_EVENT           = 1
)

type Event interface {
	getEmitter() *emitter.Client
	getType() EventType
	Handle()
}
