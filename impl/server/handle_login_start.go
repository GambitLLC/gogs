package server

import (
	"bytes"
	"fmt"
	"github.com/panjf2000/gnet"
	"gogs/api/events"
	"gogs/impl/ecs"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/packetids"
)

func (s *Server) handleLoginStart(conn gnet.Conn, pkt pk.Packet) (out []byte, err error) {
	ctx := conn.Context().(connectionContext)

	var name pk.String
	if err = pkt.Unmarshal(&name); err != nil {
		return
	}

	logger.Printf("Received login start packet from player %v", name)

	// TODO: implement encryption (if online mode, send encryption request instead of following)

	//trigger login event
	event := events.PlayerLoginData{
		Name: string(name),
		Conn: conn, // for ip address bans? consider changing to just ip
	}
	events.PlayerLoginEvent.Trigger(&event)

	if event.Result == events.LoginAllowed {
		buf := bytes.Buffer{}
		u := pk.NameToUUID(string(name)) // todo: get uuid from mojang servers
		// send login success
		buf.Write(pk.Marshal(
			packetids.LoginSuccess,
			pk.UUID(u),
			name,
		).Encode())

		player := s.createPlayer(string(name), u, conn)
		buf.Write(s.joinGamePacket(player).Encode())

		out = buf.Bytes()

		/*
			event := events.PlayerJoinData{
				Player:  api.Player(player),
				Message: fmt.Sprintf("%v has joined the game", player.Name()),
			}
			events.PlayerJoinEvent.Trigger(&event)
		*/
		s.Broadcast(fmt.Sprintf("%v has joined the game", player.Name))

		// send out new player info to everyone already online
		playerInfoPacket := clientbound.PlayerInfo{
			Action:     0,
			NumPlayers: 1,
			Players: []pk.Encodable{
				clientbound.PlayerInfoAddPlayer{
					UUID:           pk.UUID(player.UUID),
					Name:           pk.String(player.Name),
					NumProperties:  0,
					Properties:     nil,
					Gamemode:       0,
					Ping:           0,
					HasDisplayName: false,
					DisplayName:    "",
				},
			},
		}.CreatePacket()
		// TODO: spawn player should be occurring when players enter range (not join game), do logic elsewhere (tick?)
		spawnPlayerPacket := clientbound.SpawnPlayer{
			EntityID:   pk.VarInt(player.ID()),
			PlayerUUID: pk.UUID(player.UUID),
			X:          pk.Double(player.X),
			Y:          pk.Double(player.Y),
			Z:          pk.Double(player.Z),
			Yaw:        pk.Angle(player.Yaw),
			Pitch:      pk.Angle(player.Pitch),
		}.CreatePacket()

		s.broadcastPacket(playerInfoPacket, conn)
		s.broadcastPacket(spawnPlayerPacket, conn)
	} else {
		// TODO: Send disconnect packet with reason
		err = fmt.Errorf("login not allowed not yet implemented")
		_ = conn.Close()
		return
	}

	conn.SetContext(connectionContext{
		State:           playState,
		ProtocolVersion: ctx.ProtocolVersion,
	})
	return
}

func (s *Server) joinGamePacket(player *ecs.Player) pk.Packet {
	return clientbound.JoinGame{
		EntityID:     pk.Int(player.ID()),
		IsHardcore:   false,
		Gamemode:     pk.UByte(player.GameMode),
		PrevGamemode: -1,
		WorldCount:   1,
		WorldNames:   []pk.Identifier{"world"},
		DimensionCodec: pk.NBT{
			V: clientbound.DimensionCodec{
				DimensionTypes: clientbound.DimensionTypeRegistry{
					Type: "minecraft:dimension_type",
					Value: []clientbound.DimensionTypeRegistryEntry{
						{"minecraft:overworld",
							0,
							clientbound.MinecraftOverworld,
						},
					},
				},
				BiomeRegistry: clientbound.BiomeRegistry{
					Type: "minecraft:worldgen/biome",
					Value: []clientbound.BiomeRegistryEntry{
						{
							Name: "minecraft:plains",
							ID:   0,
							Element: clientbound.BiomeProperties{
								Precipitation: "none",
								Depth:         0.125,
								Temperature:   0.8,
								Scale:         0.05,
								Downfall:      0.4,
								Category:      "plains",
								Effects: clientbound.BiomeEffects{
									SkyColor:      7907327,
									WaterFogColor: 329011,
									FogColor:      12638463,
									WaterColor:    4159204,
								},
							},
						},
					},
				},
			}},
		Dimension:    pk.NBT{V: clientbound.MinecraftOverworld},
		WorldName:    "world",
		HashedSeed:   0,
		MaxPlayers:   20,
		ViewDistance: 10,
		RDI:          false,
		ERS:          false,
		IsDebug:      false,
		IsFlat:       true,
	}.CreatePacket()
}
