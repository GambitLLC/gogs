package main

import (
	"flag"
	"github.com/panjf2000/gnet"
	"gogs/impl/logger"
	"gogs/impl/server"
	"strconv"
)

func main() {
	host := flag.String("host", "127.0.0.1", "host ip")
	port := flag.Uint("port", 25565, "host port")
	flag.Parse()

	MinecraftServer := new(server.Server)
	MinecraftServer.Host = *host
	MinecraftServer.Port = uint16(*port)

	connString := "tcp://" + MinecraftServer.Host + ":" + strconv.Itoa(int(MinecraftServer.Port))

	logger.Error(
		gnet.Serve(MinecraftServer, connString, gnet.WithMulticore(true), gnet.WithTicker(true)),
	)
}
