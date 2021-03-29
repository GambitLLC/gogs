package main

import (
	"gogs/impl/server"
	"testing"
)

func TestServer(t *testing.T) {
	MinecraftServer := new(server.Server)
	MinecraftServer.Host = "localhost"
	MinecraftServer.Port = 25565

	err := MinecraftServer.Start()
	if err != nil {
		t.Error(err)
	}
}
