package listeners

import (
	emitter "github.com/emitter-io/go/v2"
	"log"
)

type LoginListener struct {
	Listener
}

func (listener LoginListener) Callback() func(*emitter.Client, emitter.Message) {
	return func(client *emitter.Client, message emitter.Message) {
		log.Println(message.Topic())
		log.Println(string(message.Payload()))
	}
}
