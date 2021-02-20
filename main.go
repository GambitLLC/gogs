package main

import (
	"github.com/panjf2000/gnet"
	"gogs/impl"
	"gogs/impl/logger"
	"strconv"
)

func main() {
	MinecraftServer := new(impl.Server)
	MinecraftServer.Host = "127.0.0.1"
	MinecraftServer.Port = 25565

	connString := "tcp://" + MinecraftServer.Host + ":" + strconv.Itoa(int(MinecraftServer.Port))

	logger.Error(
		gnet.Serve(MinecraftServer, connString, gnet.WithMulticore(true), gnet.WithTicker(true)),
	)
}
