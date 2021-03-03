package game

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"errors"
	"github.com/Tnze/go-mc/nbt"
	"io"
	"math"
)

type column struct {
	X        int
	Z        int
	Sections []chunkSection
}

type chunkSection struct {
	Y           byte
	Palette     []int32
	BlockStates compactedDataArray
	paletteMap  map[int32]uint8 // map of global palette id to palette index
}

type compactedDataArray struct {
	Data          []int64
	BitsPerValue  int
	valuesPerLong int
	capacity      int
	bitMask       int64
}

func newCompactedDataArray(bitsPervalue int, capacity int) compactedDataArray {
	v := compactedDataArray{}
	v.init(bitsPervalue, capacity)
	return v
}

func (s *compactedDataArray) init(bitsPerValue int, capacity int) {
	s.BitsPerValue = bitsPerValue
	s.valuesPerLong = 64 / bitsPerValue
	s.Data = make([]int64, int(math.Ceil(float64(capacity)/float64(s.valuesPerLong))))
	s.capacity = capacity
	s.bitMask = (1 << bitsPerValue) - 1
}

func (s compactedDataArray) get(index int) int64 {
	dataIndex := index / s.valuesPerLong
	dataShift := (index % s.valuesPerLong) * s.BitsPerValue

	return (s.Data[dataIndex] >> dataShift) & s.bitMask

	/*
		// OLD FORMAT: values can span across longs
		index *= s.BitsPerValue
		dataIndex := index >> 6
		dataShift := index & 63


		val := s.Data[dataIndex] >> dataShift
		// check if value spreads over two longs
		if dataShift + s.BitsPerValue > 64 {
			val |= s.Data[dataIndex+1] << (64 - dataShift)
		}

		return val & s.bitMask
	*/
}

func (s *compactedDataArray) set(index int, val int64) {
	dataIndex := index / s.valuesPerLong
	dataShift := (index % s.valuesPerLong) * s.BitsPerValue

	s.Data[dataIndex] &^= s.bitMask << dataShift
	s.Data[dataIndex] |= (val & s.bitMask) << dataShift

	/*
		// OLD FORMAT: values can span across longs
		index *= s.BitsPerValue
		dataIndex := index >> 6
		dataShift := index & 63

		// clear the bits needed to be set
		s.Data[dataIndex] &^= s.bitMask << dataShift
		s.Data[dataIndex] |= (val & s.bitMask) << dataShift

		// check if value spreads over two longs
		if dataShift + s.BitsPerValue > 64 {
			dataIndex += 1
			// clear the bits needed to be set
			s.Data[dataIndex] &^= int(1 << (dataShift + s.BitsPerValue - 64)) - 1
			s.Data[dataIndex] |= val >> (64 - dataShift)
		}
	*/
}

// anvilColumn is [16]anvilChunkSection
type anvilColumn struct {
	DataVersion int
	Level       struct {
		Heightmaps map[string][]int64
		Structures struct {
			References map[string][]int64
			Starts     map[string]struct {
				ID string `nbt:"id"`
			}
		}
		// Entities
		// LiquidTicks
		// PostProcessing
		Sections []anvilChunkSection
		// TileEntities
		// TileTicks
		InhabitedTime int64
		IsLightOn     byte `nbt:"isLightOn"`
		LastUpdate    int64
		Status        string
		PosX          int32 `nbt:"xPos"`
		PosZ          int32 `nbt:"zPos"`
		Biomes        []int32
	}
}

type anvilChunkSection struct {
	Palette     []anvilBlock
	Y           byte
	BlockLight  []byte
	BlockStates []int64
	SkyLight    []byte
}

type anvilBlock struct {
	Name       string
	Properties map[string]interface{}
}

// Load read column data from []byte
func (c *anvilColumn) Load(data []byte) (err error) {
	var r io.Reader = bytes.NewReader(data[1:])

	switch data[0] {
	default:
		err = errors.New("unknown compression")
	case 1:
		r, err = gzip.NewReader(r)
	case 2:
		r, err = zlib.NewReader(r)
	}

	if err != nil {
		return err
	}

	err = nbt.NewDecoder(r).Decode(c)
	return
}
