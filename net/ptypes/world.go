package ptypes

import (
	pk "gogs/net/packet"
)

type JoinGame struct {
	PlayerEntity pk.Int
	Hardcore     pk.Boolean
	Gamemode     pk.UByte
	PrevGamemode pk.UByte
	WorldCount   pk.VarInt
	WorldNames   pk.Identifier
	//DimensionCodec pk.NBT
	Dimension    pk.Int
	WorldName    pk.Identifier
	HashedSeed   pk.Long
	maxPlayers   pk.VarInt // Now ignored
	ViewDistance pk.VarInt
	RDI          pk.Boolean // Reduced Debug Info
	ERS          pk.Boolean // Enable respawn screen
	IsDebug      pk.Boolean
	IsFlat       pk.Boolean
}

func (s JoinGame) CreatePacket() pk.Packet {
	// TODO: create packetid consts
	return pk.Marshal(0x24, s.PlayerEntity, s.Hardcore, s.Gamemode,
		s.PrevGamemode, s.WorldCount, s.WorldNames, s.Dimension,
		s.WorldName, s.HashedSeed, s.maxPlayers, s.ViewDistance,
		s.RDI, s.ERS, s.IsDebug, s.IsFlat)
}
