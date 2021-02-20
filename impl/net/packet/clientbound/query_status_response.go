package clientbound

import pk "gogs/impl/net/packet"

type QueryStatusResponse struct {
	JSONResponse pk.String
}

func (p QueryStatusResponse) CreatePacket() pk.Packet {
	return pk.Marshal(0x00, p.JSONResponse)
}
