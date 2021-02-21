package packetids

const (
	// Handshake
	Handshake = 0x00

	// Status Clientbound
	StatusResponse = 0x00
	StatusPong     = 0x01

	// Status Serverbound
	StatusRequest = 0x00
	StatusPing    = 0x01

	// Login Clientbound
	LoginDisconnect    = 0x00
	EncryptionRequest  = 0x01
	LoginSuccess       = 0x02
	SetCompression     = 0x03
	LoginPluginRequest = 0x04

	// Login Serverbound
	LoginStart          = 0x00
	EncryptionResponse  = 0x01
	LoginPluginResponse = 0x02

	// Play Clientbound
	StatusEntity                     = 0x00
	SpawnExperienceOrb               = 0x01
	SpawnLivingEntity                = 0x02
	SpawnPainting                    = 0x03
	SpawnPlayer                      = 0x04
	EntityAnimation                  = 0x05
	Statistics                       = 0x06
	AcknowledgePlayerDigging         = 0x07
	BlockBreakAnimation              = 0x08
	BlockEntityData                  = 0x09
	BlockAction                      = 0x0A
	BlockChange                      = 0x0B
	BossBar                          = 0x0C
	ServerDifficulty                 = 0x0D
	ChatMessageClientbound           = 0x0E
	TabCompleteClientbound           = 0x0F
	DeclareCommands                  = 0x10
	WindowConfirmationClientbound    = 0x11
	CloseWindowClientbound           = 0x12
	WindowItems                      = 0x13
	WindowProperty                   = 0x14
	SetSlot                          = 0x15
	SetCooldown                      = 0x16
	PluginMessageClientbound         = 0x17
	NamedSoundEffect                 = 0x18
	PlayDisconnect                   = 0x19
	EntityStatus                     = 0x1A
	Explosion                        = 0x1B
	UnloadChunk                      = 0x1C
	ChangeGameState                  = 0x1D
	OpenHorseWindow                  = 0x1E
	KeepAliveClientbound             = 0x1F
	ChunkData                        = 0x20
	Effect                           = 0x21
	Particle                         = 0x22
	UpdateLight                      = 0x23
	JoinGame                         = 0x24
	MapData                          = 0x25
	TradeList                        = 0x26
	EntityPosition                   = 0x27
	EntityPositionAndRotation        = 0x28
	EntityRotation                   = 0x29
	EntityMovement                   = 0x2A
	VehicleMoveClientbound           = 0x2B
	OpenBook                         = 0x2C
	OpenWindow                       = 0x2D
	OpenSignEditor                   = 0x2E
	CraftRecipeResponse              = 0x2F
	PlayerAbilitiesClientbound       = 0x30
	CombatEvent                      = 0x31
	PlayerInfo                       = 0x32
	FacePlayer                       = 0x33
	PlayerPositionAndLookClientbound = 0x34
	UnlockRecipes                    = 0x35
	DestroyEntities                  = 0x36
	RemoveEntityEffect               = 0x37
	ResourcePackSend                 = 0x38
	Respawn                          = 0x39
	EntityHeadLook                   = 0x3A
	MultiBlockChange                 = 0x3B
	SelectAdvancementTab             = 0x3C
	WorldBorder                      = 0x3D
	Camera                           = 0x3E
	HeldItemChangeClientbound        = 0x3F
	UpdateViewPosition               = 0x40
	UpdateViewDistance               = 0x41
	SpawnPosition                    = 0x42
	DisplayScoreboard                = 0x43
	EntityMetadata                   = 0x44
	AttachEntity                     = 0x45
	EntityVelocity                   = 0x46
	EntityEquipment                  = 0x47
	SetExperience                    = 0x48
	UpdateHealth                     = 0x49
	ScoreboardObjective              = 0x4A
	SetPassengers                    = 0x4B
	Teams                            = 0x4C
	UpdateScore                      = 0x4D
	TimeUpdate                       = 0x4E
	Title                            = 0x4F
	EntitySoundEffect                = 0x50
	SoundEffect                      = 0x51
	StopSound                        = 0x52
	PlayerListHeaderAndFooter        = 0x53
	NBTQueryResponse                 = 0x54
	CollectItem                      = 0x55
	EntityTeleport                   = 0x56
	Advancements                     = 0x57
	EntityProperties                 = 0x58
	EntityEffect                     = 0x59
	DeclareRecipes                   = 0x5A
	Tags                             = 0x5B

	// Play Serverbound
	TeleportConfirm                      = 0x00
	QueryBlockNBT                        = 0x01
	QueryEntityNBT                       = 0x0D
	SetDifficulty                        = 0x02
	ChatMessageServerbound               = 0x03
	ClientStatus                         = 0x04
	ClientSettings                       = 0x05
	TabCompleteServerbound               = 0x06
	WindowConfirmationServerbound        = 0x07
	ClickWindowButton                    = 0x08
	ClickWindow                          = 0x09
	CloseWindowServerbound               = 0x0A
	PluginMessageServerServerbound       = 0x0B
	EditBook                             = 0x0C
	InteractEntity                       = 0x0E
	GenerateStructure                    = 0x0F
	KeepAliveServerbound                 = 0x10
	LockDifficulty                       = 0x11
	PlayerPosition                       = 0x12
	PlayerPositionAndRotationServerbound = 0x13
	PlayerRotation                       = 0x14
	PlayerMovement                       = 0x15
	VehicleMoveServerbound               = 0x16
	SteerBoat                            = 0x17
	PickItem                             = 0x18
	CraftRecipeRequest                   = 0x19
	PlayerAbilitiesServerbound           = 0x1A
	PlayerDigging                        = 0x1B
	EntityAction                         = 0x1C
	SteerVehicle                         = 0x1D
	SetRecipeBookState                   = 0x1E
	SetDisplayedRecipe                   = 0x1F
	NameItem                             = 0x20
	ResourcePackStatus                   = 0x21
	AdvancementTab                       = 0x22
	SelectTrade                          = 0x23
	SetBeaconEffect                      = 0x24
	HeldItemChangeServerbound            = 0x25
	UpdateCommandBlock                   = 0x26
	UpdateCommandBlockMinecart           = 0x27
	CreativeInventoryAction              = 0x28
	UpdateJigsawBlock                    = 0x29
	UpdateStructureBlock                 = 0x2A
	UpdateSign                           = 0x2B
	Animation                            = 0x2C
	Spectate                             = 0x2D
	PlayerBlockPlacement                 = 0x2E
	UseItem                              = 0x2F
)
