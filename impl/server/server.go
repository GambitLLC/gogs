package server

import (
	"bytes"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/panjf2000/gnet"
	"gogs/api/data/chat"
	api "gogs/api/game"
	implgame "gogs/impl/game"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
)

type playerMapping struct {
	uuidToPlayer map[uuid.UUID]*implgame.Player
	uuidToConn   map[uuid.UUID]gnet.Conn
	connToPlayer map[gnet.Conn]*implgame.Player
}

type Server struct {
	gnet.EventServer

	Host        string
	Port        uint16
	tickCount   uint64
	numEntities int32 // TODO: find a better way to implement entity ids?

	mu        sync.RWMutex
	playerMap *playerMapping
}

func (s *Server) CreatePlayer(name string, u uuid.UUID, conn gnet.Conn) api.Player {
	s.mu.Lock()
	defer s.mu.Unlock()

	//player, exists := s.playerMap.uuidToPlayer[u]
	//if exists {
	//	// TODO: figure out what happens to players who connect twice
	//	s.playerMap.uuidToConn[u] = conn
	//	s.playerMap.connToPlayer[conn] = player
	//	return player
	//}

	player := implgame.NewPlayer(name, u, conn, s.numEntities)
	s.numEntities += 1
	s.playerMap.uuidToPlayer[u] = player
	s.playerMap.uuidToConn[u] = conn
	s.playerMap.connToPlayer[conn] = player

	return player
}

func (s *Server) Players() []api.Player {
	s.mu.RLock()
	defer s.mu.RUnlock()

	players := make([]api.Player, 0, len(s.playerMap.uuidToPlayer))
	for _, player := range s.playerMap.uuidToPlayer {
		players = append(players, player)
	}
	return players
}

func (s *Server) PlayerFromConn(conn gnet.Conn) api.Player {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.playerMap.connToPlayer[conn]
}

func (s *Server) PlayerFromUUID(uuid uuid.UUID) api.Player {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.playerMap.uuidToPlayer[uuid]
}

func (s *Server) ConnFromUUID(uuid uuid.UUID) gnet.Conn {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.playerMap.uuidToConn[uuid]
}

func (s *Server) Broadcast(text string) {
	// TODO: figure out chat colors
	logger.Printf(text)
	msg := chat.NewMessage("§e" + text)
	pkt := clientbound.ChatMessage{
		JSONData: pk.Chat(msg.AsJSON()),
		Position: 1, // TODO: define chat positions as enum
		Sender:   pk.UUID{},
	}.CreatePacket()

	s.broadcastPacket(pkt, nil)
}

func (s *Server) broadcastPacket(pkt pk.Packet, conn gnet.Conn) {
	out := pkt.Encode()
	s.mu.RLock()
	for _, c := range s.playerMap.uuidToConn {
		if c != conn {
			_ = c.AsyncWrite(out)
		}
	}
	s.mu.RUnlock()
}

func (s *Server) Init() {
	s.playerMap = &playerMapping{
		uuidToPlayer: make(map[uuid.UUID]*implgame.Player),
		uuidToConn:   make(map[uuid.UUID]gnet.Conn),
		connToPlayer: make(map[gnet.Conn]*implgame.Player),
	}
	// TODO: set up Server initialization (world, etc)

	// TODO: PlayerLoginEvent should check if players banned/whitelisted first
}

//On Server Start - Ready to accept connections
func (s *Server) OnInitComplete(_ gnet.Server) gnet.Action {
	logger.Printf("gogs - a blazingly fast minecraft server")
	s.Init()
	logger.Printf("Server listening for connections on tcp://" + s.Host + ":" + strconv.Itoa(int(s.Port)))
	return gnet.None
}

//On Server End - Event loop and all connections closed
func (s *Server) OnShutdown(_ gnet.Server) {
	logger.Printf("Server shutting down")
}

//On Connection Opened - Player either logging in or getting status
func (s *Server) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	logger.Printf("New connection received")
	c.SetContext(connectionContext{State: handshakeState})
	return nil, gnet.None
}

//On Connection Closed - A connection has been closed
func (s *Server) OnClosed(c gnet.Conn, _ error) gnet.Action {
	logger.Printf("Connection closed")

	//clean up all the player state
	s.mu.RLock()
	player, exists := s.playerMap.connToPlayer[c]
	s.mu.RUnlock()

	if exists {
		s.mu.Lock()
		delete(s.playerMap.uuidToConn, player.UUID())
		delete(s.playerMap.uuidToPlayer, player.UUID())
		delete(s.playerMap.connToPlayer, c)
		s.mu.Unlock()

		logger.Printf("Player %v disconnected", player.Name())

		// update player info for all remaining players
		playerInfoPacket := clientbound.PlayerInfo{
			Action:     4, // TODO: create consts for action
			NumPlayers: 1,
			Players: []pk.Encodable{
				clientbound.PlayerInfoRemovePlayer{
					UUID: pk.UUID(player.UUID()),
				},
			},
		}.CreatePacket().Encode()
		// also destroy the entity for all players
		destroyEntitiesPacket := clientbound.DestroyEntities{
			Count:     1,
			EntityIDs: []pk.VarInt{pk.VarInt(player.EntityID())},
		}.CreatePacket().Encode()

		s.mu.RLock()
		for _, conn := range s.playerMap.uuidToConn {
			_ = conn.AsyncWrite(append(playerInfoPacket, destroyEntitiesPacket...))
		}
		s.mu.RUnlock()

		// TODO: trigger disconnect event
		s.Broadcast(fmt.Sprintf("%v has left the game", player.Name()))
		return gnet.None
	}

	return gnet.None
}

//On packet
func (s *Server) React(frame []byte, c gnet.Conn) ([]byte, gnet.Action) {
	out := bytes.Buffer{}
	for buf := bytes.NewReader(frame); buf.Len() > 0; {
		packet, err := pk.Decode(buf)
		if err != nil {
			logger.Printf("error decoding frame into packet: %v", err)
			return nil, gnet.None
		}

		res, err := s.handlePacket(c, *packet)
		if err != nil {
			logger.Printf("failed to handle packet %v\n got error: %v", packet, err.Error())
			return nil, gnet.None
		}
		out.Write(res)
	}

	return out.Bytes(), gnet.None
}

//On tick
func (s *Server) Tick() (delay time.Duration, action gnet.Action) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	startTime := time.Now()
	// TODO: probably game logic stuff
	if s.tickCount%100 == 0 {
		//send out keep-alive to all players
		pkt := clientbound.KeepAlive{
			ID: pk.Long(time.Now().UnixNano()),
		}.CreatePacket()
		s.broadcastPacket(pkt, nil)
	}

	//for _, p := range s.playerMap.uuidToPlayer {
	//	p.Tick(s)
	//}

	s.tickCount++
	// tick every 50 ms (20 tps)
	return time.Duration(50000000 - time.Since(startTime).Nanoseconds()), gnet.None
}
