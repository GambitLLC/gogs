package clientbound

import (
	"bytes"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/packetids"
)

type DestroyEntities struct {
	Count     pk.VarInt
	EntityIDs ids
}

func (s DestroyEntities) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.DestroyEntities, s.Count, s.EntityIDs)
}

type ids []pk.VarInt

func (a ids) Encode() []byte {
	buf := bytes.Buffer{}
	for _, v := range a {
		buf.Write(v.Encode())
	}
	return buf.Bytes()
}
