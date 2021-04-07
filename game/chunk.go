package game

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"errors"
	"github.com/Tnze/go-mc/nbt"
	"gogs/logger"
	"io"
	"math"
	"sync"
)

type column struct {
	X             int
	Z             int
	Sections      [16]*chunkSection
	BlockEntities []blockEntity

	Lock sync.RWMutex
}

// Block returns the block at the given global x, y, z. Does not check if x & z fall within the column.
func (c *column) Block(x int, y int, z int) (blockID int32) {
	c.Lock.RLock()
	defer c.Lock.RUnlock()

	sectionY := y >> 4
	if c.Sections[sectionY] == nil {
		return 0
	}

	sectionX := x % 16
	if sectionX < 0 {
		sectionX += 16
	}
	sectionZ := z % 16
	if sectionZ < 0 {
		sectionZ += 16
	}
	return c.Sections[sectionY].Block(sectionX, y%16, sectionZ)
}

// SetBlock sets the block at the given global x, y, z. Does not check if x & z fall within the column.
func (c *column) SetBlock(x int, y int, z int, blockID int32) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	sectionY := y >> 4
	if c.Sections[sectionY] == nil {
		c.Sections[sectionY] = &chunkSection{
			Y:           byte(sectionY),
			Palette:     make([]int32, 1, 16),
			BlockStates: newCompactedDataArray(4, 4096),
		}
	}
	sectionX := x % 16
	if sectionX < 0 {
		sectionX += 16
	}
	sectionZ := z % 16
	if sectionZ < 0 {
		sectionZ += 16
	}

	c.Sections[sectionY].SetBlock(sectionX, y%16, sectionZ, blockID)
}

type chunkSection struct {
	Y           byte
	Palette     []int32
	BlockStates compactedDataArray
	paletteMap  map[int32]uint8 // map of global palette id to palette index
}

// Block returns the block id at the given chunk x, chunk y, and chunk z.
func (s *chunkSection) Block(x int, y int, z int) (blockID int32) {
	if s.Palette == nil {
		return 0
	}

	paletteIndex := s.BlockStates.get(256*y + 16*z + x)
	if int(paletteIndex) >= len(s.Palette) {
		logger.Printf("paletteIndex received was greater than palette size somehow ...")
		return 0
	}

	return s.Palette[paletteIndex]
}

// SetBlock sets the block at the given chunk x, chunk y, and chunk z.
func (s *chunkSection) SetBlock(x int, y int, z int, blockID int32) {
	if s.paletteMap == nil {
		s.paletteMap = make(map[int32]uint8, 16)
	}

	// update the palette if needed
	paletteIndex, exists := s.paletteMap[blockID]
	if !exists {
		paletteIndex = uint8(len(s.Palette))
		s.paletteMap[blockID] = paletteIndex
		s.Palette = append(s.Palette, blockID)

		if len(s.Palette) > (1 << s.BlockStates.BitsPerValue) {
			s.BlockStates.upsize(s.BlockStates.BitsPerValue + 1)
		}
	}

	s.BlockStates.set(256*y+16*z+x, int64(paletteIndex))
}

type compactedDataArray struct {
	Data          []int64
	BitsPerValue  int
	valuesPerLong int
	capacity      int
	bitMask       int64
}

func newCompactedDataArray(bitsPerValue int, capacity int) compactedDataArray {
	v := compactedDataArray{}
	v.init(bitsPerValue, capacity)
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

func (s *compactedDataArray) upsize(newBitsPerValue int) {
	v := newCompactedDataArray(newBitsPerValue, s.capacity)
	for i := 0; i < s.capacity; i++ {
		v.set(i, s.get(i))
	}
	*s = v
}

// anvilColumn is [16]anvilChunkSection
type anvilColumn struct {
	DataVersion int
	Level       struct {
		//Heightmaps map[string][]int64
		//Structures struct {
		//	References map[string][]int64
		//	Starts     map[string]struct {
		//		ID string `nbt:"id"`
		//	}
		//}
		//Entities
		//LiquidTicks
		//PostProcessing
		Sections     []anvilChunkSection
		TileEntities []blockEntity
		//TileTicks
		//InhabitedTime int64
		//IsLightOn     byte `nbt:"isLightOn"`
		//LastUpdate    int64
		//Status        string
		PosX int32 `nbt:"xPos"`
		PosZ int32 `nbt:"zPos"`
		//Biomes        []int32
	}
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

type anvilChunkSection struct {
	Palette []anvilBlock
	Y       byte
	//BlockLight  []byte
	BlockStates []int64
	//SkyLight    []byte
}

type anvilBlock struct {
	Name       string
	Properties map[string]interface{}
}

type blockEntity struct {
	ID         string `nbt:"id"`
	KeepPacked byte   `nbt:"keepPacked"`
	X          int32  `nbt:"x"`
	Y          int32  `nbt:"y"`
	Z          int32  `nbt:"z"`
}
