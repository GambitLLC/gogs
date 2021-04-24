package serverbound

import (
	"bytes"
	pk "github.com/GambitLLC/gogs/net/packet"
)

type InteractEntity struct {
	EntityID pk.VarInt
	Type     pk.VarInt
	TargetX  pk.Float  // Optional, if Type == 2 (interact at)
	TargetY  pk.Float  // Optional, if Type == 2 (interact at)
	TargetZ  pk.Float  // Optional, if Type == 2 (interact at)
	Hand     pk.VarInt // Optional, if Type == 0 || Type == 2 (interact or interact at)
	Sneaking pk.Boolean
}

func (s *InteractEntity) FromPacket(packet pk.Packet) error {
	r := bytes.NewReader(packet.Data)
	for _, field := range []pk.Decodable{&s.EntityID, &s.Type} {
		if err := field.Decode(r); err != nil {
			return err
		}
	}
	if s.Type == 2 {
		for _, field := range []pk.Decodable{&s.TargetX, &s.TargetY, &s.TargetZ, &s.Hand} {
			if err := field.Decode(r); err != nil {
				return err
			}
		}
	} else if s.Type == 0 {
		if err := (&s.Hand).Decode(r); err != nil {
			return err
		}
	}

	return (&s.Sneaking).Decode(r)
}
