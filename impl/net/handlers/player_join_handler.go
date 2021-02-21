package handlers

import (
	"fmt"
	"gogs/api"
	"gogs/api/events"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
)

func PlayerJoinHandler(s api.Server) func(*events.PlayerJoinData) {
	return func(data *events.PlayerJoinData) {
		player := data.Player

		s.Broadcast(fmt.Sprintf("%v has joined the game", player.Name))

		// send the players that are already online to the person who joined
		c := s.ConnFromUUID(player.UUID)
		players := s.Players()
		playerInfoArr := make([]pk.Encodable, 0, len(players))
		for _, p := range players {
			playerInfoArr = append(playerInfoArr, clientbound.PlayerInfoAddPlayer{
				UUID:           pk.UUID(p.UUID),
				Name:           pk.String(p.Name),
				NumProperties:  pk.VarInt(0),
				Properties:     nil,
				Gamemode:       pk.VarInt(0),
				Ping:           pk.VarInt(0),
				HasDisplayName: false,
				DisplayName:    "",
			})
		}
		c.AsyncWrite(clientbound.PlayerInfo{
			Action:     0,
			NumPlayers: pk.VarInt(len(players)),
			Players:    playerInfoArr,
		}.CreatePacket().Encode())

		// send the player who just joined to everyone
		for _, p := range players {
			c := s.ConnFromUUID(p.UUID)
			err := c.AsyncWrite(clientbound.PlayerInfo{
				Action:     0,
				NumPlayers: 1,
				Players: []pk.Encodable{
					clientbound.PlayerInfoAddPlayer{
						UUID:           pk.UUID(player.UUID),
						Name:           pk.String(player.Name),
						NumProperties:  pk.VarInt(0),
						Properties:     nil,
						Gamemode:       pk.VarInt(0),
						Ping:           pk.VarInt(0),
						HasDisplayName: false,
						DisplayName:    "",
					},
				},
			}.CreatePacket().Encode())
			if err != nil {
				logger.Printf("error sending player info, %w", err)
			}
		}
	}
}
