package game

import (
	"github.com/Tnze/go-mc/save/region"
	"math"
)

const BITS_PER_BLOCK = 15

type WorldBlock struct {
	Name string
	X    int32
	Y    int32
	Z    int32
}

type WorldChunk struct {
	Blocks [256][16][16]WorldBlock
}

func Nibble4(arr []byte, index int) byte {
	if index%2 == 0 {
		return arr[index/2] & 0x0F
	} else {
		return arr[index/2] >> 4 & 0x0F
	}
}

func bits(bs []byte) []int {
	r := make([]int, len(bs)*8)
	for i, b := range bs {
		for j := 0; j < 8; j++ {
			r[i*8+j] = int(b >> uint(7-j) & 0x01)
		}
	}
	return r
}

func (w *World) GetWorldChunk(x int, z int) (*WorldChunk, error) {
	if w.worldChunkMap == nil {
		w.worldChunkMap = make(map[int]map[int]*WorldChunk)
	}

	if w.worldChunkMap[x] == nil {
		w.worldChunkMap[x] = make(map[int]*WorldChunk)
	}

	if w.worldChunkMap[x][z] == nil {
		worldChunk := WorldChunk{}

		columnChunk, _ := w.GetChunk(x, z)
		_y := int32(0)
		_t := 0
		for _, section := range columnChunk.Level.Sections {
			_break := false

			if section.BlockStates == nil {
				for i := 0; i < 4096; i++ {
					_x := int32(i % 16)
					_z := int32((i / 16) % 16)

					_x += (columnChunk.Level.PosX * 16)
					_z += (columnChunk.Level.PosY * 16)

					worldChunk.Blocks[_y][_z][_x] = WorldBlock{
						Name: "minecraft:air",
						X:    _x,
						Y:    _y,
						Z:    _z,
					}

					if i%256 == 0 {
						_y++
					}
				}
				continue
			}

			bitsPerBlock := int64(math.Log2(float64(len(section.Palette))))
			if bitsPerBlock < 4 {
				bitsPerBlock = 4
			}

			_x := int32(0)
			_z := int32(0)
			for _, blockState := range section.BlockStates {
				if _break {
					break
				}

				/*
					_x := int32(i % 16)
					_z := int32((i / 16) % 16)

					_x += (columnChunk.Level.PosX * 16)
					_z += (columnChunk.Level.PosY * 16)

					for j := 0; j < int(64/bitsPerBlock); j++ {
						blockId := blockState & ((0b1 << bitsPerBlock) - 1)
						blockState = blockState >> bitsPerBlock

						worldChunk.Blocks[_y][_z][_x] = WorldBlock{
							Name: section.Palette[blockId].Name,
							X:  _x,
							Y:  _y,
							Z:  _z,
						}
					}

					if i != 0 && i % 256 == 0 {
						_y++
					}
				*/

				for j := 0; j < int(64/bitsPerBlock); j++ {

					var block uint64 = uint64(blockState) & uint64(uint64((0b1<<bitsPerBlock)-1)<<((15-j)*4))
					block = block >> ((15 - j) * 4)

					worldChunk.Blocks[_y][_z][_x] = WorldBlock{
						Name: section.Palette[block].Name,
						X:    _x + (columnChunk.Level.PosX * 16),
						Y:    _y,
						Z:    _z + (columnChunk.Level.PosY * 16),
					}

					_t++
					_x++

					if _t%16 == 0 {
						_x = 0
						_z++
					}
					if _t%256 == 0 {
						_z = 0
						_y++
					}
					if _t == 4096 {
						_t = 0
						_break = true
						break
					}
				}
			}
			_y++
		}
		w.worldChunkMap[x][z] = &worldChunk
		return &worldChunk, nil
	}

	return w.worldChunkMap[x][z], nil
}

type World struct {
	columnMap     map[int]map[int]*Column
	worldChunkMap map[int]map[int]*WorldChunk
}

func (w *World) GetChunk(x int, z int) (*Column, error) {
	if w.columnMap == nil {
		w.columnMap = make(map[int]map[int]*Column)
	}

	if w.columnMap[x] == nil {
		w.columnMap[x] = make(map[int]*Column)
	}

	if w.columnMap[x][z] == nil {
		r, err := region.Open("./test_world/region/r.0.0.mca")
		if err != nil {
			return nil, err
		}
		defer r.Close()

		data, err := r.ReadSector(x, z)
		if err != nil {
			return nil, err
		}

		var c Column
		err = c.Load(data)
		if err != nil {
			return nil, err
		}

		w.columnMap[x][z] = &c
		return &c, nil
	}

	return w.columnMap[x][z], nil
}
