package main

import (
	"flag"
	"gogs/impl/logger"
	"gogs/impl/server"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
)

func main() {
	host := flag.String("host", "127.0.0.1", "host ip")
	port := flag.Uint("port", 25565, "host port")
	profile := flag.Bool("profile", false, "enable pprof")
	flag.Parse()

	if *profile {
		runtime.SetBlockProfileRate(1)
		runtime.SetMutexProfileFraction(1)
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
		log.Println("pprof http server listening on http://localhost:6060/debug/pprof/")
	}

	MinecraftServer := new(server.Server)
	MinecraftServer.Host = *host
	MinecraftServer.Port = uint16(*port)

	logger.Error(MinecraftServer.Start())
}
