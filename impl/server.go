package impl

import (
	"bytes"
	"github.com/google/uuid"
	"gogs/api/events"
	"gogs/impl/logger"
	"gogs/impl/net/packet/clientbound"
	"strconv"
	"time"

	"github.com/panjf2000/gnet"
	"gogs/api/game"
	plists "gogs/impl/net/listeners"
	pk "gogs/impl/net/packet"
)

type playerMapping struct {
	uuidToPlayer map[uuid.UUID]*game.Player
	uuidToConn   map[uuid.UUID]gnet.Conn
	connToUUID   map[gnet.Conn]uuid.UUID
}

type Server struct {
	Host string
	Port uint16
	gnet.EventServer
	tickCount uint64
	playerMap *playerMapping
}

func (s *Server) CreatePlayer(name string, uuid uuid.UUID, conn gnet.Conn) *game.Player {
	player, exists := s.playerMap.uuidToPlayer[uuid]
	if exists {
		// TODO: figure out what happens to players who connect twice
		s.playerMap.uuidToConn[uuid] = conn
		return player
	}
	player = &game.Player{
		UUID: uuid,
		Name: name,
		Position: game.Position{
			X: 0,
			Y: 2,
			Z: 0,
		},
		Rotation: game.Rotation{
			Yaw:   0,
			Pitch: 0,
		},
		SpawnPosition: game.Position{
			X: 0,
			Y: 2,
			Z: 0,
		},
	}
	s.playerMap.uuidToPlayer[uuid] = player
	s.playerMap.uuidToConn[uuid] = conn
	s.playerMap.connToUUID[conn] = uuid

	return player
}

func (s *Server) Players() []*game.Player {
	players := make([]*game.Player, 0, len(s.playerMap.uuidToPlayer))
	for _, player := range s.playerMap.uuidToPlayer {
		players = append(players, player)
	}
	return players
}

func (s *Server) PlayerFromConn(conn gnet.Conn) *game.Player {
	return s.playerMap.uuidToPlayer[s.playerMap.connToUUID[conn]]
}

func (s *Server) Init() {
	s.playerMap = &playerMapping{
		uuidToPlayer: make(map[uuid.UUID]*game.Player),
		uuidToConn:   make(map[uuid.UUID]gnet.Conn),
		connToUUID:   make(map[gnet.Conn]uuid.UUID),
	}
	// TODO: set up Server initialization (world, etc)

	// TODO: Move all net listeners into another file

	// TODO: PlayerLoginEvent should check if players banned/whitelisted first
	events.PlayerLoginEvent.RegisterNet(func(event *events.PlayerLoginData) {
		// send login success
		if event.Result == events.LoginAllowed {
			err := event.Conn.AsyncWrite(pk.Marshal(
				0x02,
				pk.UUID(event.Player.UUID),
				pk.String(event.Player.Name),
			).Encode())
			if err != nil {
				logger.Printf("error sending login success, %w", err)
			}
		} else {
			// TODO: send kick message
		}
	})

	events.PlayerJoinEvent.RegisterNet(func(data *events.PlayerJoinData) {
		player := data.Player
		c := s.playerMap.uuidToConn[player.UUID]
		// send the players that are already online
		players := make([]pk.Encodable, 0, len(s.playerMap.uuidToPlayer))
		for _, p := range s.playerMap.uuidToPlayer {
			players = append(players, clientbound.PlayerInfoAddPlayer{
				UUID:           pk.UUID(p.UUID),
				Name:           pk.String(p.Name),
				NumProperties:  pk.VarInt(0),
				Properties:     nil,
				Gamemode:       pk.VarInt(0),
				Ping:           pk.VarInt(0),
				HasDisplayName: false,
				DisplayName:    "",
			})
		}
		c.AsyncWrite(clientbound.PlayerInfo{
			Action:     0,
			NumPlayers: pk.VarInt(len(players)),
			Players:    players,
		}.CreatePacket().Encode())

		// send the player who just joined to everyone else
		for _, c := range s.playerMap.uuidToConn {
			err := c.AsyncWrite(clientbound.PlayerInfo{
				Action:     0,
				NumPlayers: 1,
				Players: []pk.Encodable{
					clientbound.PlayerInfoAddPlayer{
						UUID:           pk.UUID(player.UUID),
						Name:           pk.String(player.Name),
						NumProperties:  pk.VarInt(0),
						Properties:     nil,
						Gamemode:       pk.VarInt(0),
						Ping:           pk.VarInt(0),
						HasDisplayName: false,
						DisplayName:    "",
					},
				},
			}.CreatePacket().Encode())
			if err != nil {
				logger.Printf("error sending player info, %w", err)
			}
		}
	})

	events.PlayerChatEvent.RegisterNet(func(data *events.PlayerChatData) {
		msg := clientbound.ChatMessage{
			JSONData: pk.Chat(data.AsJSON()),
			Position: 0,
			Sender:   pk.UUID(data.Player.UUID),
		}.CreatePacket().Encode()
		for _, p := range data.Recipients {
			c := s.playerMap.uuidToConn[p.UUID]
			c.AsyncWrite(msg)
		}
	})
}

//On Server Start - Ready to accept connections
func (s *Server) OnInitComplete(svr gnet.Server) gnet.Action {
	logger.Printf("gogs - a blazingly fast minecraft server")
	s.Init()
	logger.Printf("Server listening for connections on tcp://" + s.Host + ":" + strconv.Itoa(int(s.Port)))
	return gnet.None
}

//On Server End - Event loop and all connections closed
func (s *Server) OnShutdown(svr gnet.Server) {
	logger.Printf("Server shutting down")
}

//On Connection Opened - Player either logging in or getting status
func (s *Server) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	logger.Printf("New connection received")
	c.SetContext(plists.HandshakePacketListener{S: s})
	return nil, gnet.None
}

//On Connection Closed - A connection has been closed
func (s *Server) OnClosed(c gnet.Conn, err error) gnet.Action {
	logger.Printf("Connection closed")

	//clean up all the player state
	delete(s.playerMap.uuidToConn, s.playerMap.connToUUID[c])
	delete(s.playerMap.uuidToPlayer, s.playerMap.connToUUID[c])
	delete(s.playerMap.connToUUID, c)

	return gnet.None
}

//On packet
func (s *Server) React(frame []byte, c gnet.Conn) (o []byte, action gnet.Action) {
	packet, err := pk.Decode(bytes.NewReader(frame))
	if err != nil {
		logger.Printf("error: %w", err)
		return nil, gnet.None
	}
	logger.Printf("packet came in: %v", *packet)

	plist := c.Context().(plists.PacketListener)
	out, err := plist.HandlePacket(c, packet)
	if err != nil {
		logger.Printf("failed to handle packet, got error: %w", err)
		return nil, gnet.None
	}

	c.AsyncWrite(out)

	return nil, gnet.None
}

//On tick
func (s *Server) Tick() (delay time.Duration, action gnet.Action) {
	startTime := time.Now()
	// TODO: probably game logic stuff
	if s.tickCount%100 == 0 {
		//send out keep-alive to all players
		for _, c := range s.playerMap.uuidToConn {
			c.AsyncWrite(clientbound.KeepAlive{
				ID: pk.Long(time.Now().UnixNano()),
			}.CreatePacket().Encode())
		}
	}

	s.tickCount++
	return time.Duration(50000000 - time.Since(startTime).Nanoseconds()), gnet.None
}
