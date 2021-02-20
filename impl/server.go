package impl

import (
	"bytes"
	"github.com/google/uuid"
	"gogs/api/events"
	"log"
	"time"

	"github.com/panjf2000/gnet"
	"gogs/api/game"
	plists "gogs/impl/net/listeners"
	pk "gogs/net/packet"
)

type playerMapping struct {
	all []*game.Player
	uuidToPlayer map[uuid.UUID]*game.Player
	uuidToConn map[uuid.UUID]gnet.Conn
}

func (m *playerMapping) CreatePlayer(name string, uuid uuid.UUID) bool {
	_, exists := m.uuidToPlayer[uuid]
	if exists {
		return false
	}
	player := game.Player{
		UUID: uuid,
		Name: name,
	}
	m.all = append(m.all, &player)
	m.uuidToPlayer[uuid] = &player
	return true
}

func (m *playerMapping) Add(data *events.PlayerLoginData) {
	m.uuidToConn[data.UUID] = data.Conn
}

type Server struct {
	gnet.EventServer

	players *playerMapping
}

func (s *Server) Load() {
	s.players = &playerMapping{
		uuidToPlayer: make(map[uuid.UUID]*game.Player),
		uuidToConn:   make(map[uuid.UUID]gnet.Conn),
	}
	// TODO: set up Server initialization (world, etc)

	// TODO: PlayerLoginEvent should check if players banned/whitelisted first
	events.PlayerLoginEvent.Register(s.players.Add)
	events.PlayerLoginEvent.RegisterNet(func(data *events.PlayerLoginData) {
		// send login success
		if data.Result == events.LoginAllowed {
			err := data.Conn.SendTo(pk.Marshal(
				0x02,
				pk.UUID(data.UUID),
				pk.String(data.Name),
			).Encode())
			if err != nil {
				log.Printf("error sending login success, %w", err)
			}
		} else {
			// TODO: send kick message
		}
	})
}

//On Server Start - Ready to accept connections
func (s *Server) OnInitComplete(svr gnet.Server) gnet.Action {
	log.Printf("Server listening for connections")
	s.Load()
	log.Printf("Server ready")
	return gnet.None
}

//On Server End - Event loop and all connections closed
func (s *Server) OnShutdown(svr gnet.Server) {
	log.Printf("Server shutting down")
}

//On Connection Opened - Player either logging in or getting status
func (s *Server) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	log.Printf("New connection received")
	c.SetContext(plists.HandshakePacketListener{S: s})
	return nil, gnet.None
}

//On Connection Closed - A connection has been closed
func (s *Server) OnClosed(c gnet.Conn, err error) gnet.Action {
	log.Printf("Connection closed")
	return gnet.None
}

//On packet
func (s *Server) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	packet, err := pk.Decode(bytes.NewReader(frame))
	if err != nil {
		// TODO: Should connections really be closed on error?
		log.Printf("error: %w", err)
		return nil, gnet.Close
	}
	log.Printf("packet came in: %v", packet)

	plist := c.Context().(plists.PacketListener)
	if err := plist.HandlePacket(c, packet); err != nil {
		log.Printf("failed to handle packet, got error: %w", err)
		return nil, gnet.Close
	}

	return nil, gnet.None
}

//On tick
func (s *Server) Tick() (delay time.Duration, action gnet.Action) {
	startTime := time.Now()

	// TODO: probably game logic stuff

	return time.Duration(50000000 - time.Since(startTime).Nanoseconds()), gnet.None
}