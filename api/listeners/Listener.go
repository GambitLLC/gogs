package listeners

import emitter "github.com/emitter-io/go/v2"

type EventListener interface {
	Callback() func(*emitter.Client, emitter.Message)
	GetName() string
	GetLinkName() string
	GetLink() *emitter.Link
	SetLink(link *emitter.Link, linkName string)
}
