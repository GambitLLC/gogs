package server

import (
	"fmt"
	"github.com/panjf2000/gnet"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/serverbound"
)

func (s *Server) handleEntityAction(conn gnet.Conn, pkt pk.Packet) ([]byte, error) {
	in := serverbound.EntityAction{}
	if err := in.FromPacket(pkt); err != nil {
		return nil, err
	}

	player := s.PlayerFromConn(conn)
	switch in.ActionID {
	case 0: // start sneaking
		s.broadcastPacket(clientbound.EntityMetadata{
			EntityID: pk.VarInt(player.EntityID()),
			Metadata: []clientbound.MetadataField{
				{Index: 6, Type: 18, Value: pk.VarInt(5)}, // SNEAKING
				{Index: 0xFF},
			},
		}.CreatePacket(), conn)
	case 1: // stop sneaking
		s.broadcastPacket(clientbound.EntityMetadata{
			EntityID: pk.VarInt(player.EntityID()),
			Metadata: []clientbound.MetadataField{
				{Index: 6, Type: 18, Value: pk.VarInt(0)}, // STANDING
				{Index: 0xFF},
			},
		}.CreatePacket(), conn)
	case 2: // leave bed
	case 3: // start sprinting
	case 4: // stop sprinting
	case 5: // start jump with horse
	case 6: // stop jump with horse
	case 7: // open horse inventory
	case 8: // start flying with elytra
	default:
		return nil, fmt.Errorf("invalid action id %d", in.ActionID)
	}

	return nil, nil
}
