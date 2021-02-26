package game

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"errors"
	"github.com/Tnze/go-mc/nbt"
	"io"
)

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
		PosY          int32 `nbt:"yPos"`
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
