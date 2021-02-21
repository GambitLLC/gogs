package handlers

import (
	"gogs/api"
	"gogs/api/events"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
)

func PlayerLoginHandler(s api.Server) func(*events.PlayerLoginData) {
	return func(event *events.PlayerLoginData) {
		// send login success
		if event.Result == events.LoginAllowed {
			err := event.Conn.AsyncWrite(pk.Marshal(
				0x02,
				pk.UUID(event.Player.UUID),
				pk.String(event.Player.Name),
			).Encode())
			if err != nil {
				logger.Printf("error sending login success, %w", err)
			}
		} else {
			// TODO: send kick message
		}
	}
}
