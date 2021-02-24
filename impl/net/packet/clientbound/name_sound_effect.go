package clientbound

import (
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/packetids"
)

type NamedSoundEffect struct {
	SoundName       pk.Identifier
	SoundCategory   pk.VarInt
	EffectPositionX pk.Int   // https://wiki.vg/Data_types#Fixed-point_numbers
	EffectPositionY pk.Int   // Fixed point with 3 bits for fractional part
	EffectPositionZ pk.Int   // Equal to EffectX multiplied by 8
	Volume          pk.Float // 1 is 100%, can be greater
	Pitch           pk.Float // Between 0.5 and 2.0 on Notchian clients
}

func (s NamedSoundEffect) CreatePacket() pk.Packet {
	return pk.Marshal(
		packetids.NamedSoundEffect,
		s.SoundName,
		s.SoundCategory,
		s.EffectPositionX,
		s.EffectPositionY,
		s.EffectPositionZ,
		s.Volume,
		s.Pitch,
	)
}
