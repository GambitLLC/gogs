package main

import (
	"github.com/panjf2000/gnet"
	"gogs/impl"
	"log"
)

func main() {
	MinecraftServer := new(impl.Server)
	log.Fatal(
		gnet.Serve(MinecraftServer, "tcp://0.0.0.0:25565", gnet.WithMulticore(true)),
	)
}
