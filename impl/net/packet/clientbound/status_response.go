package clientbound

import (
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/packetids"
)

type StatusResponse struct {
	JSONResponse pk.String
}

func (p StatusResponse) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.StatusResponse, p.JSONResponse)
}
