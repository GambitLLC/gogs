package handlers

import (
	"encoding/json"
	"github.com/panjf2000/gnet"
	"gogs/api"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
)

func StatusRequest(c gnet.Conn, _ *pk.Packet, s api.Server) error {
	logger.Printf("Received status request packet")
	// TODO: fill response with info from server
	resp := response{
		Version: version{
			Name:     "gogs 1.16.5",
			Protocol: 754,
		},
		Description: description{
			Text: "gogs - a blazingly fast minecraft server",
		},
		Players: players{
			Max:    20,
			Online: len(s.Players()),
			Sample: nil,
		},
		Favicon: "",
	}

	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}

	packet := clientbound.StatusResponse{
		JSONResponse: pk.String(jsonBytes),
	}.CreatePacket().Encode()

	return c.AsyncWrite(packet)
}

type response struct {
	Version version `json:"version"`

	Description description `json:"description"`

	Players players `json:"players"`
	Favicon string  `json:"favicon,omitempty"`
}

type version struct {
	Name     string `json:"name"`
	Protocol int    `json:"protocol"`
}

type description struct {
	Text string `json:"text"`
}

type players struct {
	Max    int           `json:"max"`
	Online int           `json:"online"`
	Sample []interface{} `json:"sample,omitempty"`
}
