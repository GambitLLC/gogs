package listeners

import (
	"errors"
	"fmt"
	"github.com/panjf2000/gnet"
	"gogs/api"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/serverbound"
)

type StatusPacketListener struct {
	S               api.Server
	protocolVersion int32
}

func (listener StatusPacketListener) HandlePacket(c gnet.Conn, p *pk.Packet) ([]byte, error) {
	//respond with query pong packet
	switch p.ID {
	//QueryResponsePacket
	case 0x00:
		outPacket := clientbound.StatusResponse{
			JSONResponse: `{"description":{"text":"gogs - a blazingly fast minecraft server"},"players":{"max":20,"online":0},"version":{"name":"gogs 1.16.5","protocol":754}}`,
		}.CreatePacket().Encode()

		return outPacket, nil

	//QueryPongPacket
	case 0x01:
		ping := serverbound.QueryStatusPing{}
		if err := ping.FromPacket(p); err != nil {
			return nil, err
		}

		return clientbound.StatusPong{
			Payload: ping.Payload,
		}.CreatePacket().Encode(), nil

	default:
		return nil, errors.New(fmt.Sprintf("Illegal packet id recieved: %02X", p.ID))
	}
}
