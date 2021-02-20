package packet

import (
	"bytes"
	"errors"
)

type Packet struct {
	ID   int32
	Data []byte
}

// Marshal creates a Packet from the given id and fields
func Marshal(id int32, fields ...Encodable) (p Packet) {
	p.ID = id
	for _, field := range fields {
		p.Data = append(p.Data, field.Encode()...)
	}
	return
}

// Unmarshal fills the fields from the Packet Data
func (p Packet) Unmarshal(fields ...Decodable) error {
	r := bytes.NewReader(p.Data)
	for _, field := range fields {
		err := field.Decode(r)
		if err != nil {
			return err
		}
	}
	return nil
}

// Encode will format the Packet into the byte array to be sent
func (p Packet) Encode() (bi []byte) {
	d := append(VarInt(p.ID).Encode(), p.Data...)
	length := VarInt(len(d)).Encode()
	bi = append(length, d...)
	return
}

// Decode will create a Packet from the given byte array
func Decode(r PacketReader) (*Packet, error) {
	var length VarInt
	if err := length.Decode(r); err != nil {
		return nil, err
	}

	if length < 1 {
		return nil, errors.New("packet is too short")
	}

	// read the entire packet first
	bi := make([]byte, length)
	if _, err := r.Read(bi); err != nil {
		return nil, err
	}

	// TODO: decompress

	br := bytes.NewBuffer(bi)
	var id VarInt
	if err := id.Decode(br); err != nil {
		return nil, errors.New("failed to read packet ID")
	}

	return &Packet{ID: int32(id), Data: br.Bytes()}, nil
}
