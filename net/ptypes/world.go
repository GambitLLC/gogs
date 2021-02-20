package ptypes

import (
	pk "gogs/net/packet"
)

type worldNames []pk.Identifier

func (a worldNames) Encode() []byte {
	var bs []byte
	for _, v := range a {
		bs = append(bs, v.Encode()...)
	}
	return bs
}

type DimensionCodec struct {
	DimensionTypes DimensionTypeRegistry `nbt:"minecraft:dimension_type"`
	BiomeRegistry BiomeRegistry `nbt:"minecraft:worldgen/biome"`
}

type DimensionTypeRegistry struct {
	Type string `nbt:"type"`
	Value []DimensionTypeRegistryEntry `nbt:"value"`
}

type DimensionTypeRegistryEntry struct {
	Name string `nbt:"name"`
	ID int32 `nbt:"id"`
	Element DimensionType `nbt:"element"`
}

type DimensionType struct {
	PiglinSafe byte `nbt:"piglin_safe"`
	Natural byte `nbt:"natural"`
	AmbientLight float32 `nbt:"ambient_light"`
	Infiniburn string `nbt:"infiniburn"`
	RespawnAnchorWorks byte `nbt:"respawn_anchor_works"`
	HasSkylight byte `nbt:"has_skylight"`
	BedWorks byte `nbt:"bed_works"`
	Effects string `nbt:"effects"`
	HasRaids byte `nbt:"has_raids"`
	LogicalHeight int32 `nbt:"logical_height"`
	CoordinateScale float32 `nbt:"coordinate_scale"`
	Ultrawarm byte `nbt:"ultrawarm"`
	HasCeiling byte `nbt:"has_ceiling"`
}

var MinecraftOverworld = DimensionType{
	PiglinSafe:         0,
	Natural:            1,
	AmbientLight:       1.0,
	Infiniburn:         "",
	RespawnAnchorWorks: 0,
	HasSkylight:        1,
	BedWorks:           1,
	Effects:            "minecraft:overworld",
	HasRaids:           0,
	LogicalHeight:      0,
	CoordinateScale:    1.0,
	Ultrawarm:          0,
	HasCeiling:         0,
}

type BiomeRegistry struct {
	Type string `nbt:"type"`
	Value []BiomeRegistryEntry `nbt:"value"`
}

type BiomeRegistryEntry struct {
	Name string `nbt:"name"`
	ID int32 `nbt:"id"`
	Element BiomeProperties `nbt:"element"`
}

type BiomeProperties struct {
	Precipitation string `nbt:"precipitation"`
	Depth float32 `nbt:"depth"`
	Temperature float32 `nbt:"temperature"`
	Scale float32 `nbt:"scale"`
	Downfall float32 `nbt:"downfall"`
	Category string `nbt:"category"`
	Effects BiomeEffects `nbt:"effects"`
}

type BiomeEffects struct {
	SkyColor int32 `nbt:"sky_color"`
	WaterFogColor int32 `nbt:"water_fog_color"`
	FogColor int32 `nbt:"fog_color"`
	WaterColor int32 `nbt:"water_color"`
}

type JoinGame struct {
	PlayerEntity pk.Int
	Hardcore     pk.Boolean
	Gamemode     pk.UByte
	PrevGamemode pk.Byte
	WorldCount   pk.VarInt
	WorldNames   worldNames	// Array of Identifiers
	DimensionCodec pk.NBT
	Dimension    pk.NBT
	WorldName    pk.Identifier
	HashedSeed   pk.Long
	MaxPlayers   pk.VarInt // Now ignored
	ViewDistance pk.VarInt
	RDI          pk.Boolean // Reduced Debug Info
	ERS          pk.Boolean // Enable respawn screen
	IsDebug      pk.Boolean
	IsFlat       pk.Boolean
}

func (s JoinGame) CreatePacket() pk.Packet {
	// TODO: create packetid consts
	return pk.Marshal(0x24, s.PlayerEntity, s.Hardcore, s.Gamemode,
		s.PrevGamemode, s.WorldCount, s.WorldNames, s.DimensionCodec,
		s.Dimension, s.WorldName, s.HashedSeed, s.MaxPlayers, s.ViewDistance,
		s.RDI, s.ERS, s.IsDebug, s.IsFlat)
}
