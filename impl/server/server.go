package server

import (
	"encoding/json"
	"github.com/google/uuid"
	"gogs/api/data/chat"
	"gogs/impl/data"
	"gogs/impl/ecs"
	"gogs/impl/game"
	"gogs/impl/logger"
	"gogs/impl/net"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"io/ioutil"
	"log"
	"os"
	"sync"
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

	Host         string
	Port         uint16
	tickCount    uint64
	shuttingDown bool

	playerMap *playerMapping
	entityMap map[uint64]interface{}
	world     *game.World
}

func (s *Server) Start() error {
	s.init()
	return s.listen()
}

func (s *Server) init() {
	if err := s.loadSettings(); err != nil {
		panic(err)
	}

	s.playerMap = &playerMapping{
		uuidToPlayer: make(map[uuid.UUID]*ecs.Player),
		connToPlayer: make(map[net.Conn]*ecs.Player),
	}
	s.entityMap = make(map[uint64]interface{})

	s.world = &game.World{
		WorldName: s.WorldName,
	}
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

func (s *Server) listen() error {
	l, err := net.NewListener(s.Host, int(s.Port))
	if err != nil {
		return err
	}

	log.Printf("Server listening for connections on tcp://%s:%d", s.Host, s.Port)

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		go s.handleHandshake(conn)
	}
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

	spawnPos := ecs.PositionComponent{
		X: 0,
		Y: 90,
		Z: 0,
	}
	player = &ecs.Player{
		BasicEntity:         ecs.NewEntity(data.ProtocolID("minecraft:entity_type", "minecraft:player")),
		PositionComponent:   spawnPos,
		HealthComponent:     ecs.HealthComponent{Health: 20},
		FoodComponent:       ecs.FoodComponent{Food: 20, Saturation: 0},
		ConnectionComponent: ecs.ConnectionComponent{Connection: conn},
		InventoryComponent: ecs.InventoryComponent{
			Inventory: make([]pk.Slot, 46), // https://wiki.vg/Inventory#Player_Inventory
		},
		SpawnPosition: spawnPos,
		GameMode:      s.GameMode,
		UUID:          u,
		Name:          name,
	}

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

func (s *Server) playerFromEntityID(id uint64) *ecs.Player {
	// todo: consider creating a map
	// todo: should be getEntity() and not just for players
	s.playerMap.Lock.RLock()
	defer s.playerMap.Lock.RUnlock()

	for _, p := range s.playerMap.uuidToPlayer {
		if p.ID() == id {
			return p
		}
	}
	return nil
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
