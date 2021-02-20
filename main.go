package main

import (
	"github.com/panjf2000/gnet"
	"gogs/impl"
	"gogs/impl/logger"
)

func main() {
	MinecraftServer := new(impl.Server)
	logger.Error(
		gnet.Serve(MinecraftServer, "tcp://0.0.0.0:25565", gnet.WithMulticore(true), gnet.WithTicker(true)),
	)
}
