package game

import (
	"fmt"
	"github.com/Tnze/go-mc/save/region"
	"gogs/impl/data"
	"math"
)

type World struct {
	WorldName string
	columnMap map[int]map[int]*column
}

func (w *World) SetBlock(x int, y int, z int, blockID int32) {
	w.Column(x>>4, z>>4).SetBlock(x, y, z, blockID)
}

func (w *World) Column(x int, z int) *column {
	if w.columnMap == nil {
		w.columnMap = make(map[int]map[int]*column)
	}

	if w.columnMap[x] == nil {
		w.columnMap[x] = make(map[int]*column)
	}

	if w.columnMap[x][z] == nil {
		w.loadRegion(x>>5, z>>5)
	}

	return w.columnMap[x][z]
}

// loadRegion loads all chunks in the region into the mapping.
func (w *World) loadRegion(regionX int, regionZ int) {
	r, rErr := region.Open(fmt.Sprintf("./%s/region/r.%d.%d.mca", w.WorldName, regionX, regionZ))
	if rErr != nil {
		// store empty columns if region file couldn't be opened
		for x := 0; x < 32; x += 1 {
			for z := 0; z < 32; z += 1 {
				w.storeColumn(regionX<<5+x, regionZ<<5+z, nil)
			}
		}
		return
	}
	defer r.Close()

	for x := 0; x < 32; x += 1 {
		for z := 0; z < 32; z += 1 {
			sector, err := r.ReadSector(z, x)
			if err != nil {
				w.storeColumn(regionX<<5+x, regionZ<<5+z, nil)
				continue
			}

			var c anvilColumn
			err = c.Load(sector)
			if err != nil {
				w.storeColumn(regionX<<5+x, regionZ<<5+z, nil)
				continue
			}

			w.storeColumn(regionX<<5+x, regionZ<<5+z, &c)
		}
	}
}

func (w *World) storeColumn(x int, z int, c *anvilColumn) {
	val := column{
		X:        x,
		Z:        z,
		Sections: [16]*chunkSection{},
	}

	if c != nil {
		val.BlockEntities = c.Level.TileEntities
		for _, section := range c.Level.Sections {
			// ignore empty sections
			if section.Palette == nil {
				continue
			}

			paletteLength := len(section.Palette)
			palette := make([]int32, paletteLength)
			paletteMap := make(map[int32]uint8, paletteLength)

			for i, block := range section.Palette {
				id := data.BlockStateID(block.Name, block.Properties)
				palette[i] = id
				paletteMap[id] = uint8(i)
			}

			// don't store empty air chunks (anvil file seems to store them)
			if len(palette) == 1 && palette[0] == 0 {
				continue
			}

			bitsPerBlock := int(math.Ceil(math.Log2(float64(len(section.Palette)))))
			if bitsPerBlock < 4 {
				bitsPerBlock = 4
			}

			// copy over directly from anvil format
			blockStates := newCompactedDataArray(bitsPerBlock, 4096)
			for i, block := range section.BlockStates {
				blockStates.Data[i] = block
			}

			val.Sections[section.Y] = &chunkSection{
				Y:           section.Y,
				Palette:     palette,
				paletteMap:  paletteMap,
				BlockStates: blockStates,
			}
		}
	}

	if w.columnMap[x] == nil {
		w.columnMap[x] = make(map[int]*column)
	}
	w.columnMap[x][z] = &val
}
