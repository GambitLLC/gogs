package clientbound

import pk "gogs/impl/net/packet"

type DeclareRecipes struct {
	NumRecipes pk.VarInt
	Recipes    recipes
}

func (s DeclareRecipes) CreatePacket() pk.Packet {
	// TODO: create packetid consts
	return pk.Marshal(0x5A, s.NumRecipes, s.Recipes)
}

type recipes []recipe

func (a recipes) Encode() []byte {
	var bs []byte
	for _, v := range a {
		bs = append(bs, v.Encode()...)
	}
	return bs
}

// TODO: create recipe struct
type recipe struct {
}

func (s recipe) Encode() []byte {
	return nil
}
