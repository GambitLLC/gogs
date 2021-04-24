package clientbound

import (
	pk "github.com/GambitLLC/gogs/net/packet"
	"github.com/GambitLLC/gogs/net/packet/packetids"
)

type Respawn struct {
	Dimension        pk.NBT
	WorldName        pk.Identifier
	HashedSeed       pk.Long
	Gamemode         pk.UByte
	PreviousGamemode pk.UByte
	IsDebug          pk.Boolean
	IsFlat           pk.Boolean
	CopyMetadata     pk.Boolean
}

func (s Respawn) CreatePacket() pk.Packet {
	return pk.Marshal(
		packetids.Respawn,
		s.Dimension,
		s.WorldName,
		s.HashedSeed,
		s.Gamemode,
		s.PreviousGamemode,
		s.IsDebug,
		s.IsFlat,
		s.CopyMetadata,
	)
}
