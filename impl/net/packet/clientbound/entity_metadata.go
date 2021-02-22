package clientbound

import (
	"bytes"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/packetids"
)

type EntityMetadata struct {
	EntityID pk.VarInt
	Metadata metadataArray
}

func (s EntityMetadata) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.EntityMetadata, s.EntityID, s.Metadata)
}

type metadataArray []MetadataField

func (a metadataArray) Encode() []byte {
	buf := bytes.Buffer{}
	for _, v := range a {
		buf.Write(v.Encode())
	}
	return buf.Bytes()
}

// https://wiki.vg/Entity_metadata#Entity_Metadata_Format
type MetadataField struct {
	Index pk.UByte
	Type  pk.VarInt    // Optional, present if Index != 0xff
	Value pk.Encodable // Optional, present if Index != 0xff, type depends on Type
}

func (s MetadataField) Encode() []byte {
	if s.Index == 0xff {
		return s.Index.Encode()
	} else {
		buf := bytes.Buffer{}
		buf.Write(s.Index.Encode())
		buf.Write(s.Type.Encode())
		buf.Write(s.Value.Encode())
		return buf.Bytes()
	}
}
