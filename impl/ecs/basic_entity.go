package ecs

import (
	"sync/atomic"
)

var idCounter uint64

type Entity interface {
	ID() uint64
	Type() int32
}

type BasicEntity struct {
	id         uint64
	entityType int32
}

func NewEntity(entityType int32) BasicEntity {
	return BasicEntity{
		id:         atomic.AddUint64(&idCounter, 1),
		entityType: entityType,
	}
}

func (s BasicEntity) ID() uint64 {
	return s.id
}

func (s BasicEntity) Type() int32 {
	return s.entityType
}
