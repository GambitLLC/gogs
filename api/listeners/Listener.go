package listeners

import emitter "github.com/emitter-io/go/v2"

type Listener interface {
	Callback() func(*emitter.Client, emitter.Message)
}
