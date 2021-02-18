package bedgg

import (
	"bedgg-server/api/listeners"
	"fmt"
	emitter "github.com/emitter-io/go/v2"
	"log"
	"strconv"
)

const (
	CHANNEL_KEY  = "ShoEzjf7DSEXdFNYM9UHbcH7frmdMdUG"
	CHANNEL_NAME = "game-server/"
)

//Monitor when new players reveal their presence
func RegisterNewPresenceHandler(client *emitter.Client, cb func(_client *emitter.Client, event emitter.PresenceEvent)) {
	client.OnPresence(cb)
}

//Subscribe to a new channel and register a callback
func RegisterNewSubscriber(client *emitter.Client, listener listeners.Listener) error {
	// Subscribe to sdk-integration-test channel
	log.Println("[emitter] subscribing to: " + CHANNEL_NAME)
	err := client.Subscribe(CHANNEL_KEY, CHANNEL_NAME, listener.Callback())
	return err
}

func NewEmitter(host string, port int16) (*emitter.Client, error) {
	// Create the client and connect to the broker
	_port := strconv.Itoa(int(port))
	c, err := emitter.Connect("tcp://"+host+":"+_port, func(_ *emitter.Client, msg emitter.Message) {
		fmt.Printf("[emitter] -> [B] received: '%s' topic: '%s'\n", msg.Payload(), msg.Topic())
	})
	if err != nil {
		return nil, err
	}

	return c, nil
}
