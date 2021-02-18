package game

import "github.com/panjf2000/gnet"

type Player struct {
	uuid      string
	username  string
	conn      gnet.Conn
	emitterID string
}

func (p Player) GetUUID() string {
	return p.uuid
}
