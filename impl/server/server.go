package server

import (
	"encoding/json"
	"github.com/google/uuid"
	"gogs/api/data/chat"
	"gogs/impl/ecs"
	"gogs/impl/game"
	"gogs/impl/logger"
	"gogs/impl/net"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/packetids"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

type playerMapping struct {
	Lock         sync.RWMutex
	uuidToPlayer map[uuid.UUID]*ecs.Player
	connToPlayer map[net.Conn]*ecs.Player
}

type serverSettings struct {
	WorldName    string
	GameMode     uint8
	ViewDistance uint8
	MaxPlayers   uint8
}

type Server struct {
	serverSettings

	Host string
	Port uint16

	listener  *net.MCListener
	ticker    *time.Ticker
	tickCount uint64

	wg       sync.WaitGroup
	shutdown chan interface{}

	playerMap playerMapping
	entityMap map[uint64]ecs.Entity
	world     game.World
}

func (s *Server) Start() error {
	s.init()
	s.tickLoop()
	return s.listen()
}

func (s *Server) init() {
	if err := s.loadSettings(); err != nil {
		panic(err)
	}

	s.playerMap = playerMapping{
		uuidToPlayer: make(map[uuid.UUID]*ecs.Player),
		connToPlayer: make(map[net.Conn]*ecs.Player),
	}
	s.entityMap = make(map[uint64]ecs.Entity)

	s.world = game.World{
		WorldName: s.WorldName,
	}

	// make channels
	s.shutdown = make(chan interface{})
}

func (s *Server) loadSettings() error {
	// default server settings
	s.serverSettings = serverSettings{
		WorldName:    "test_world",
		GameMode:     0,
		ViewDistance: 10,
		MaxPlayers:   20,
	}

	// Open our jsonFile
	jsonFile, err := os.Open("./settings.json")
	if err != nil {
		// create the settings file if it doesn't exist
		jsonFile, err = os.Create("./settings.json")
		if err != nil {
			panic(err)
		}
		defer jsonFile.Close()

		byteValue, err := json.Marshal(s.serverSettings)
		if err != nil {
			return err
		}
		_, err = jsonFile.Write(byteValue)
		return err
	}

	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(byteValue, &s.serverSettings)
	return err
}

// tickLoop runs the main server ticking in a separate goroutine.
func (s *Server) tickLoop() {
	s.ticker = time.NewTicker(50 * time.Millisecond)

	go func() {
		for {
			select {
			case <-s.shutdown:
				return
			case <-s.ticker.C:
				// do tick stuff
				s.tickCount++
			}
		}
	}()
}

func (s *Server) listen() (err error) {
	s.listener, err = net.NewListener(s.Host, int(s.Port))
	if err != nil {
		return
	}

	log.Printf("Server listening for connections on tcp://%s:%d", s.Host, s.Port)

	s.wg.Add(1)
	defer s.wg.Done()

	var conn net.Conn
	for {
		conn, err = s.listener.Accept()
		if err != nil {
			select {
			case <-s.shutdown:
				return nil
			default:
				return err
			}
		}

		go func() {
			s.wg.Add(1)
			s.handleConnection(conn)
			s.wg.Done()
		}()
	}
}

func (s *Server) stop() {
	logger.Printf("Shutting down...")
	close(s.shutdown)

	s.ticker.Stop()

	// send disconnect packet to all connections before closing listener
	s.playerMap.Lock.Lock()
	for conn := range s.playerMap.connToPlayer {
		_ = conn.WritePacket(pk.Marshal(
			packetids.PlayDisconnect,
			pk.Chat(chat.NewMessage("Server shut down").AsJSON()),
		))
		_ = conn.Close()
	}
	s.playerMap.Lock.Unlock()

	// wait for all connection handlers to finish
	s.wg.Wait()

	// cannot close listener earlier b/c it will close all connections
	_ = s.listener.Close()
}

func (s *Server) createPlayer(name string, u uuid.UUID, conn net.Conn) *ecs.Player {
	s.playerMap.Lock.Lock()
	defer s.playerMap.Lock.Unlock()

	player, exists := s.playerMap.uuidToPlayer[u]
	if exists {
		// TODO: figure out what happens to players who connect twice
		s.playerMap.connToPlayer[conn] = player
		player.Connection = conn
		return player
	}

	player = ecs.NewPlayer()
	player.Connection = conn
	player.GameMode = s.GameMode
	player.UUID = u
	player.Name = name

	// send some starting stacks for now
	player.Inventory[36] = pk.Slot{
		Present:   true,
		ItemID:    1,
		ItemCount: 64,
		NBT:       pk.NBT{},
	}
	player.Inventory[38] = pk.Slot{
		Present:   true,
		ItemID:    1,
		ItemCount: 1,
		NBT:       pk.NBT{},
	}
	player.Inventory[37] = pk.Slot{
		Present:   true,
		ItemID:    3,
		ItemCount: 64,
		NBT:       pk.NBT{},
	}

	s.playerMap.uuidToPlayer[u] = player
	s.playerMap.connToPlayer[conn] = player
	s.entityMap[player.ID()] = player

	return player
}

func (s *Server) playerFromConn(conn net.Conn) *ecs.Player {
	s.playerMap.Lock.RLock()
	defer s.playerMap.Lock.RUnlock()

	return s.playerMap.connToPlayer[conn]
}

func (s *Server) entityFromID(id uint64) ecs.Entity {
	return s.entityMap[id]
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

func (s *Server) broadcastPacket(pkt pk.Packet, exception net.Conn) {
	s.playerMap.Lock.RLock()
	for conn := range s.playerMap.connToPlayer {
		if conn != exception {
			_ = conn.WritePacket(pkt)
		}
	}
	s.playerMap.Lock.RUnlock()
}
