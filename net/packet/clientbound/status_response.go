package clientbound

import (
	pk "github.com/GambitLLC/gogs/net/packet"
	"github.com/GambitLLC/gogs/net/packet/packetids"
)

type StatusResponse struct {
	JSONResponse pk.String
}

func (p StatusResponse) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.StatusResponse, p.JSONResponse)
}
