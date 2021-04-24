package clientbound

import (
	pk "github.com/GambitLLC/gogs/net/packet"
	"github.com/GambitLLC/gogs/net/packet/packetids"
)

type UpdateHealth struct {
	Health         pk.Float
	Food           pk.VarInt
	FoodSaturation pk.Float
}

func (s UpdateHealth) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.UpdateHealth, s.Health, s.Food, s.FoodSaturation)
}
