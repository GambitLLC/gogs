package packet

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"io"
	"math"

	"github.com/Tnze/go-mc/nbt"
	"github.com/google/uuid"
)

// Field represents a field in a packet. Can be Encoded & Decoded
type Field interface {
	Encodable
	Decodable
}

type Encodable interface {
	Encode() []byte
}

type Decodable interface {
	Decode(r Reader) error
}

// Reader is used for decoding Fields
type Reader interface {
	io.ByteReader
	io.Reader
}

type (
	Boolean    bool
	Byte       int8
	UByte      uint8
	Short      int16
	UShort     uint16
	Int        int32
	Long       int64
	Float      float32
	Double     float64
	String     string
	Chat       = String
	Identifier = String
	VarInt     int32
	VarLong    int64
	Slot       struct {
		Present   Boolean
		ItemID    VarInt
		ItemCount Byte
		NBT       NBT
	}
	NBT struct {
		V interface{}
	}
	Position struct {
		X, Y, Z int32
	}
	Angle     int8
	UUID      uuid.UUID
	ByteArray []byte
)

func ReadBytes(r Reader, n int) (bs []byte, err error) {
	bs = make([]byte, n)
	for i := 0; i < n; i++ {
		bs[i], err = r.ReadByte()
		if err != nil {
			return
		}
	}
	return
}

func (b Boolean) Encode() []byte {
	if b {
		return []byte{0x01}
	}
	return []byte{0x00}
}

func (b *Boolean) Decode(r Reader) error {
	v, err := r.ReadByte()
	if err != nil {
		return err
	}

	*b = v != 0
	return nil
}

func (b Byte) Encode() []byte {
	return []byte{byte(b)}
}

func (b *Byte) Decode(r Reader) error {
	v, err := r.ReadByte()
	if err != nil {
		return err
	}
	*b = Byte(v)
	return nil
}

func (ub UByte) Encode() []byte {
	return []byte{byte(ub)}
}

func (ub *UByte) Decode(r Reader) error {
	v, err := r.ReadByte()
	if err != nil {
		return err
	}
	*ub = UByte(v)
	return nil
}

func (s Short) Encode() (bs []byte) {
	v := uint16(s)
	bs = make([]byte, 2)
	binary.BigEndian.PutUint16(bs, v)
	return
}

func (s *Short) Decode(r Reader) error {
	bs, err := ReadBytes(r, 2)
	if err != nil {
		return err
	}

	*s = Short(binary.BigEndian.Uint16(bs))
	return nil
}

func (us UShort) Encode() (bs []byte) {
	v := uint16(us)
	bs = make([]byte, 2)
	binary.BigEndian.PutUint16(bs, v)
	return
}

func (us *UShort) Decode(r Reader) error {
	bs, err := ReadBytes(r, 2)
	if err != nil {
		return err
	}

	*us = UShort(binary.BigEndian.Uint16(bs))
	return nil
}

func (n Int) Encode() (bs []byte) {
	v := uint32(n)
	bs = make([]byte, 4)
	binary.BigEndian.PutUint32(bs, v)
	return
}

func (n *Int) Decode(r Reader) error {
	bs, err := ReadBytes(r, 4)
	if err != nil {
		return err
	}

	*n = Int(binary.BigEndian.Uint32(bs))
	return nil
}

func (n Long) Encode() (bs []byte) {
	v := uint64(n)
	bs = make([]byte, 8)
	binary.BigEndian.PutUint64(bs, v)
	return
}

func (n *Long) Decode(r Reader) error {
	bs, err := ReadBytes(r, 8)
	if err != nil {
		return err
	}

	*n = Long(binary.BigEndian.Uint64(bs))
	return nil
}

func (f Float) Encode() (bs []byte) {
	v := math.Float32bits(float32(f))
	bs = make([]byte, 4)
	binary.BigEndian.PutUint32(bs, v)
	return
}

func (f *Float) Decode(r Reader) error {
	bs, err := ReadBytes(r, 4)
	if err != nil {
		return err
	}

	*f = Float(math.Float32frombits(binary.BigEndian.Uint32(bs)))
	return nil
}

func (d Double) Encode() (bs []byte) {
	v := math.Float64bits(float64(d))
	bs = make([]byte, 8)
	binary.BigEndian.PutUint64(bs, v)
	return
}

func (d *Double) Decode(r Reader) error {
	bs, err := ReadBytes(r, 8)
	if err != nil {
		return err
	}

	*d = Double(math.Float64frombits(binary.BigEndian.Uint64(bs)))
	return nil
}

func (s String) Encode() []byte {
	return append(VarInt(len(s)).Encode(), []byte(s)...)
}

func (s *String) Decode(r Reader) error {
	var length VarInt
	if err := length.Decode(r); err != nil {
		return err
	}

	bs, err := ReadBytes(r, int(length))
	if err != nil {
		return err
	}

	*s = String(bs)
	return nil
}

func (v VarInt) Encode() (vs []byte) {
	n := uint32(v)
	for {
		b := n & 0b01111111
		n >>= 7
		if n != 0 {
			b |= 0b10000000
		}

		vs = append(vs, byte(b))

		if n == 0 {
			break
		}
	}

	return
}

func (v *VarInt) Decode(r Reader) error {
	var res int32

	for i := 0; ; i++ {
		b, err := r.ReadByte()
		if err != nil {
			return err
		}

		res |= int32(b&0b01111111) << (7 * i)

		if i >= 5 {
			return errors.New("VarInt is too big")
		}

		if (b & 0b10000000) == 0 {
			break
		}
	}

	*v = VarInt(res)
	return nil
}

func (s Slot) Encode() []byte {
	var bs bytes.Buffer

	bs.Write(s.Present.Encode())
	if s.Present {
		bs.Write(s.ItemID.Encode())
		bs.Write(s.ItemCount.Encode())
		bs.Write(s.NBT.Encode())
	}

	return bs.Bytes()
}

func (s *Slot) Decode(r Reader) error {
	if err := s.Present.Decode(r); err != nil {
		return err
	}

	if s.Present {
		if err := s.ItemID.Decode(r); err != nil {
			return err
		}
		if err := s.ItemCount.Decode(r); err != nil {
			return err
		}
		if s.NBT.V != nil {
			if err := s.NBT.Decode(r); err != nil {
				return err
			}
		}
	}

	return nil
}

func (n NBT) Encode() []byte {
	var bs bytes.Buffer

	if n.V != nil {
		if err := nbt.NewEncoder(&bs).Encode(n.V); err != nil {
			panic(err)
		}
	} else {
		return []byte{nbt.TagEnd}
	}

	return bs.Bytes()
}

func (n *NBT) Decode(r Reader) error {
	return nbt.NewDecoder(r).Decode(n.V)
}

func (p Position) Encode() (bs []byte) {
	var v uint64
	v = (uint64(p.X&0x3FFFFFF) << 38) | (uint64(p.Z&0x3FFFFFF) << 12) | uint64(p.Y&0xFFF)
	bs = make([]byte, 8)
	binary.BigEndian.PutUint64(bs, v)
	return
}

func (p *Position) Decode(r Reader) error {
	bs, err := ReadBytes(r, 8)
	if err != nil {
		return err
	}

	v := binary.BigEndian.Uint64(bs)
	p.X = int32(int64(v) >> 38)
	p.Y = int32(v & 0xFFF)
	p.Z = int32(int64(v) << 26 >> 38)
	return nil
}

func (a Angle) Encode() []byte {
	return []byte{byte(a)}
}

func (u UUID) Encode() []byte {
	return u[:]
}

func (u *UUID) Decode(r Reader) error {
	_, err := r.Read((*u)[:])
	return err
}

// NameToUUID return the UUID from player name in offline mode
// TODO: implement yggdrasil authentication
func NameToUUID(name string) uuid.UUID {
	var version = 3
	h := md5.New()
	h.Reset()
	h.Write([]byte("OfflinePlayer:" + name))
	s := h.Sum(nil)
	var id uuid.UUID
	copy(id[:], s)
	id[6] = (id[6] & 0x0f) | uint8((version&0xf)<<4)
	id[8] = (id[8] & 0x3f) | 0x80 // RFC 4122 variant
	return id
}

func (b ByteArray) Encode() []byte {
	return append(VarInt(len(b)).Encode(), b...)
}

func (b *ByteArray) Decode(r Reader) error {
	var length VarInt
	if err := length.Decode(r); err != nil {
		return err
	}

	*b = make([]byte, length)
	_, err := r.Read(*b)
	return err
}
