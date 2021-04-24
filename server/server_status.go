package server

import (
	"encoding/json"
	"github.com/GambitLLC/gogs/logger"
	pk "github.com/GambitLLC/gogs/net/packet"
	"github.com/GambitLLC/gogs/net/packet/clientbound"
)

func (s *Server) statusPacket() (pk.Packet, error) {
	logger.Printf("Received status request packet")
	resp := response{
		Version: version{
			Name:     "gogs 1.16.5",
			Protocol: 754,
		},
		Description: description{
			Text: "gogs - a blazingly fast minecraft server",
		},
		Players: players{
			Max:    int(s.MaxPlayers),
			Online: len(s.playerMap.connToPlayer),
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
	}.CreatePacket()

	return packet, nil
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
