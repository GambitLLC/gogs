package main

import (
	"bytes"
	"log"
	"time"

	"gogs/api/listeners"
	io "gogs/io"
	plists "gogs/net/listeners"
	pk "gogs/net/packet"

	"github.com/panjf2000/gnet"
)

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
	c.SetContext(plists.HandshakePacketListener())
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
