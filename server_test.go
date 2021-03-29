package main

import (
	"github.com/panjf2000/gnet"
	"gogs/impl/server"
	"strconv"
	"testing"
)

func TestServer(t *testing.T) {
	MinecraftServer := new(server.Server)
	MinecraftServer.Host = "localhost"
	MinecraftServer.Port = 25565

	connString := "tcp://" + MinecraftServer.Host + ":" + strconv.Itoa(int(MinecraftServer.Port))

	err := gnet.Serve(MinecraftServer, connString, gnet.WithMulticore(false), gnet.WithTicker(true))
	if err != nil {
		t.Error(err)
	}
}
