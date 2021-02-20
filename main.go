package main

import (
	"log"
	"time"

	"github.com/panjf2000/gnet"
	"gogs/api/listeners"
	"gogs/impl"
	io "gogs/io"
)



func main() {
	go func() {
		echo := new(impl.Server)
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
