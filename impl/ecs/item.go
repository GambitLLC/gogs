package ecs

import pk "gogs/impl/net/packet"

type ItemEntity struct {
	BasicEntity
	PositionComponent
	VelocityComponent
	Item pk.Slot
}
