package impl

import (
	"bytes"
	"github.com/google/uuid"
	"gogs/api/events"
	"gogs/impl/net/packet/clientbound"
	"log"
	"time"

	"github.com/panjf2000/gnet"
	"gogs/api/game"
	plists "gogs/impl/net/listeners"
	pk "gogs/impl/net/packet"
)

type playerMapping struct {
	all []*game.Player
	uuidToPlayer map[uuid.UUID]*game.Player
	uuidToConn map[uuid.UUID]gnet.Conn
}

type Server struct {
	gnet.EventServer

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
	}
	s.playerMap.all = append(s.playerMap.all, player)
	s.playerMap.uuidToPlayer[uuid] = player
	s.playerMap.uuidToConn[uuid] = conn
	return player
}

func (s *Server) Init() {
	s.playerMap = &playerMapping{
		uuidToPlayer: make(map[uuid.UUID]*game.Player),
		uuidToConn:   make(map[uuid.UUID]gnet.Conn),
	}
	// TODO: set up Server initialization (world, etc)

	// TODO: PlayerLoginEvent should check if players banned/whitelisted first
	events.PlayerLoginEvent.RegisterNet(func(data *events.PlayerLoginData) {
		c := s.playerMap.uuidToConn[data.Player.UUID]
		// send login success
		if data.Result == events.LoginAllowed {
			err := c.SendTo(pk.Marshal(
				0x02,
				pk.UUID(data.Player.UUID),
				pk.String(data.Player.Name),
			).Encode())
			if err != nil {
				log.Printf("error sending login success, %w", err)
			}
		} else {
			// TODO: send kick message
		}
	})


	events.PlayerJoinEvent.RegisterNet(func(data *events.PlayerJoinData) {
		player := data.Player
		for _, c := range s.playerMap.uuidToConn {
			//err := c.SendTo(clientbound.PlayerInfo{
			//	Action:     0,
			//	NumPlayers: 1,
			//	Players:     []pk.Encodable{
			//		clientbound.PlayerInfoAddPlayer{
			//			UUID: 			pk.UUID(player.UUID),
			//			Name:           pk.String(player.Name),
			//			NumProperties:  pk.VarInt(0),
			//			Properties:     nil,
			//			Gamemode:       pk.VarInt(0),
			//			Ping:           pk.VarInt(0),
			//			HasDisplayName: false,
			//			DisplayName:    "",
			//		},
			//	},
			//}.Encode())
			//if err != nil {
			//	log.Printf("error sending player info, %w", err)
			//}
			log.Print(c)
			log.Print(clientbound.PlayerInfo{
					Action:     0,
					NumPlayers: 1,
					Players:     []pk.Encodable{
						clientbound.PlayerInfoAddPlayer{
							UUID: 			pk.UUID(player.UUID),
							Name:           pk.String(player.Name),
							NumProperties:  pk.VarInt(0),
							Properties:     nil,
							Gamemode:       pk.VarInt(0),
							Ping:           pk.VarInt(0),
							HasDisplayName: false,
							DisplayName:    "",
						},
					},
				})
		}
	})

}

//On Server Start - Ready to accept connections
func (s *Server) OnInitComplete(svr gnet.Server) gnet.Action {
	log.Printf("Server listening for connections")
	s.Init()
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
		log.Printf("error: %w", err)
		return nil, gnet.None
	}
	log.Printf("packet came in: %v", packet)

	plist := c.Context().(plists.PacketListener)
	if err := plist.HandlePacket(c, packet); err != nil {
		log.Printf("failed to handle packet, got error: %w", err)
		return nil, gnet.None
	}

	return nil, gnet.None
}

//On tick
func (s *Server) Tick() (delay time.Duration, action gnet.Action) {
	startTime := time.Now()

	// TODO: probably game logic stuff

	return time.Duration(50000000 - time.Since(startTime).Nanoseconds()), gnet.None
}