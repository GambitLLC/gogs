package clientbound

import pk "gogs/impl/net/packet"

type recipeInfo []pk.Identifier

func (a recipeInfo) Encode() []byte {
	var bs []byte
	for _, v := range a {
		bs = append(bs, v.Encode()...)
	}
	return bs
}

type DeclareRecipes struct {
	NumRecipes	pk.VarInt
	Recipe		recipeInfo	// TODO: array of multiple types (identifier/optional)
}

func (s DeclareRecipes) CreatePacket() pk.Packet {
	// TODO: create packetid consts
	return pk.Marshal(0x5A, s.NumRecipes, s.Recipe)
}