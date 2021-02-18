package main

import (
	"bytes"
	"errors"
	"log"
	"time"

	"gogs/api/listeners"
	io "gogs/io"
	pk "gogs/net/packet"
	//ptypes "gogs/net/ptypes"

	"github.com/panjf2000/gnet"
)

type CONNECTION_STATE int

const (
	HANDSHAKING CONNECTION_STATE = 0
	STATUS                       = 1
	LOGIN                        = 2
	PLAY                         = 3
)

type Context struct {
	State        CONNECTION_STATE
	HandlePacket func(gnet.Conn, *pk.Packet) ([]byte, error)
}

func handleHandshake(c gnet.Conn, p *pk.Packet) ([]byte, error) {
	if p.ID != 0 {
		return nil, errors.New("handshake expects Packet ID 0")
	}

	var (
		protocolVersion pk.VarInt
		address         pk.String
		port            pk.UShort
		nextState       pk.VarInt
	)

	err := p.Unmarshal(&protocolVersion, &address, &port, &nextState)
	if err != nil {
		return nil, err
	}

	c.SetContext(Context{CONNECTION_STATE(nextState), handleLoginStart})

	return nil, nil
}

func handleLoginStart(c gnet.Conn, p *pk.Packet) (out []byte, err error) {
	if p.ID != 0 {
		return nil, errors.New("login start expects Packet ID 0")
	}

	var (
		name pk.String
	)

	err = p.Unmarshal(&name)
	if err != nil {
		return
	}

	log.Printf("received login from player %v", name)

	if len(name) > 16 {
		// TODO: define packetid consts and use them
		// send disconnect
		out = pk.Marshal(0x00, pk.Chat("username too long")).Encode()
		err = errors.New("username too long")
		return
	}

	/*
		// TODO: send encryption request
		out = pk.Marshal(
			0x01,
			pk.String(""),    // Server ID
			pk.VarInt(1),    // public key length
			pk.ByteArray([]byte("s")), // public key in bytes
			pk.VarInt(1),    // verify token length
			pk.ByteArray([]byte("s")), // verify token in bytes
		).Encode()
	*/

	c.SetContext(Context{PLAY, nil})
	// send login success (offline mode for now)
	out = pk.Marshal(
		0x02,
		pk.UUID(pk.NameToUUID(string(name))), // UUID
		pk.String(name),                      // Username
	).Encode()

	// also send out join game
	/*
		out = append(out, ptypes.JoinGame{
			PlayerEntity: 0,
			Hardcore:     false,
			Gamemode:     0,
			PrevGamemode: 0,
			WorldCount:   1,
			WorldNames:   "",
			Dimension:    0,
			WorldName:    "",
			HashedSeed:   0,
			ViewDistance: 0,
			RDI:          false,
			ERS:          false,
			IsDebug:      false,
			IsFlat:       false,
		}.CreatePacket().Encode()...)
	*/
	out = append(out, pk.Marshal(0x24,
		pk.Int(0),            // EntityID
		pk.UByte(1),          // Gamemode
		pk.Int(0),            // Dimension
		pk.Long(0),           // HashedSeed
		pk.UByte(20),         // MaxPlayer
		pk.String("default"), // LevelType
		pk.VarInt(15),        // View Distance
		pk.Boolean(false),    // Reduced Debug Info
		pk.Boolean(true),     // Enable respawn screen
	).Encode()...)

	return
}

// bed.gg server
type server struct {
	*gnet.EventServer
}

//On Server Start - Ready to accept connections
func (s *server) OnInitComplete(svr gnet.Server) gnet.Action {
	log.Printf("Server started listening for connections")
	return gnet.None
}

//On Server End - Event loop and all connections closed
func (s *server) OnShutdown(svr gnet.Server) {
	log.Printf("Server shutting down")
}

//On Connection Opened - Player either logging in or getting status
func (s *server) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	log.Printf("New connection received")
	c.SetContext(Context{HANDSHAKING, handleHandshake})
	return nil, gnet.None
}

//On Connection Closed - A connection has been closed
func (s *server) OnClosed(c gnet.Conn, err error) gnet.Action {
	log.Printf("Connection closed")
	return gnet.None
}

//On packet
func (s *server) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	packet, err := pk.Decode(bytes.NewReader(frame))
	if err != nil {
		log.Printf("error: %w", err)
		return nil, gnet.None
	}

	ctx := c.Context().(Context)
	log.Printf("packet came in during state %v: %v", ctx.State, packet)
	// TODO: State isn't really necessary since handler func is overwritten
	switch ctx.State {
	case HANDSHAKING:
		fallthrough
	case LOGIN:
		out, err = ctx.HandlePacket(c, packet)
		if err != nil {
			log.Println(err)
			action = gnet.Close
			return
		}
	default:
		log.Printf("Unhandled connection state %v", ctx.State)
		out = nil
	}

	action = gnet.None
	return
}

//On tick
func (s *server) Tick() (delay time.Duration, action gnet.Action) {
	startTime := time.Now()

	// TODO: probably game logic stuff

	return time.Duration(50000000 - time.Since(startTime).Nanoseconds()), gnet.None
}

func main() {
	go func() {
		echo := new(server)
		log.Fatal(
			gnet.Serve(echo, "tcp://0.0.0.0:25565", gnet.WithMulticore(true)),
		)
	}()


	c, err := io.NewEmitter("127.0.0.1", 8080)
	if err != nil {
		log.Printf("Fatal error occured: %v", err.Error())
		return
	}

	err = io.RegisterNewSubscriber(c, &listeners.LoginListener{})
	if err != nil {
		log.Fatal(err)
		return
	}

	time.Sleep(time.Second * 2)
	c.Publish(io.CHANNEL_KEY, io.CHANNEL_NAME, "hello, world")

	select {}
}
