package listeners

import (
	emitter "github.com/emitter-io/go/v2"
	"log"
)

type LoginEventListener struct {
	name     string
	link     *emitter.Link
	linkName string
}

func NewLoginEventListener() *LoginEventListener {
	return &LoginEventListener{"LoginEventListener", nil, ""}
}

func (listener LoginEventListener) Callback() func(*emitter.Client, emitter.Message) {
	return func(client *emitter.Client, message emitter.Message) {
		logger.Println(message.Topic())
		logger.Println(string(message.Payload()))
	}
}

func (listener LoginEventListener) GetName() string {
	return listener.name
}

func (listener LoginEventListener) GetLinkName() string {
	return listener.linkName
}

func (listener *LoginEventListener) SetLink(link *emitter.Link, linkName string) {
	listener.link = link
	listener.linkName = linkName
}

func (listener LoginEventListener) GetLink() *emitter.Link {
	return listener.link
}
