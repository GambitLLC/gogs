package impl

import (
	"bytes"
	"github.com/google/uuid"
	"gogs/api"
	"gogs/api/data/chat"
	"gogs/api/events"
	"gogs/impl/logger"
	"gogs/impl/net/handlers"
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
	api.Server
	gnet.EventServer

	Host      string
	Port      uint16
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

func (s Server) Players() []*game.Player {
	players := make([]*game.Player, 0, len(s.playerMap.uuidToPlayer))
	for _, player := range s.playerMap.uuidToPlayer {
		players = append(players, player)
	}
	return players
}

func (s Server) PlayerFromConn(conn gnet.Conn) *game.Player {
	return s.playerMap.uuidToPlayer[s.playerMap.connToUUID[conn]]
}

func (s Server) PlayerFromUUID(uuid uuid.UUID) *game.Player {
	return s.playerMap.uuidToPlayer[uuid]
}

func (s Server) ConnFromUUID(uuid uuid.UUID) gnet.Conn {
	return s.playerMap.uuidToConn[uuid]
}

func (s Server) Broadcast(text string) {
	// TODO: figure out chat colors
	msg := chat.NewMessage("§e" + text)
	pkt := clientbound.ChatMessage{
		JSONData: pk.Chat(msg.AsJSON()),
		Position: 1, // TODO: define chat positions as enum
		Sender:   pk.UUID{},
	}.CreatePacket().Encode()
	for _, c := range s.playerMap.uuidToConn {
		c.AsyncWrite(pkt)
	}
}

func (s *Server) Init() {
	s.playerMap = &playerMapping{
		uuidToPlayer: make(map[uuid.UUID]*game.Player),
		uuidToConn:   make(map[uuid.UUID]gnet.Conn),
		connToUUID:   make(map[gnet.Conn]uuid.UUID),
	}
	// TODO: set up Server initialization (world, etc)

	// TODO: PlayerLoginEvent should check if players banned/whitelisted first

	// TODO: Move all net listeners into another file
	events.PlayerLoginEvent.RegisterNet(handlers.PlayerLoginHandler(s))
	events.PlayerJoinEvent.RegisterNet(handlers.PlayerJoinHandler(s))
	events.PlayerChatEvent.RegisterNet(handlers.PlayerChatHandler(s))
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
