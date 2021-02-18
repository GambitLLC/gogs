package api

import "gogs/api/game"

type Manager struct {
	playerMap     map[string]game.Player
	onlinePlayers int
}

func (es *Manager) init() {
	es.playerMap = make(map[string]game.Player)
}

func (es *Manager) addPlayer(player game.Player) {
	es.playerMap[player.GetUUID()] = player
}
