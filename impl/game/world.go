package game

import (
	"fmt"
	"github.com/Tnze/go-mc/save/region"
	"gogs/impl/data"
	"math"
)

type World struct {
	anvilColumnMap map[int]map[int]*anvilColumn
	columnMap      map[int]map[int]*column
}

func (w *World) SetBlock(x int, y int, z int, blockID int32) {
	w.GetColumn(x>>16, z>>16).SetBlock(x, y, z, blockID)
}

func (w *World) GetColumn(x int, z int) *column {
	if w.columnMap == nil {
		w.columnMap = make(map[int]map[int]*column)
	}

	if w.columnMap[x] == nil {
		w.columnMap[x] = make(map[int]*column)
	}

	if w.columnMap[x][z] == nil {
		val := column{
			X:        x,
			Z:        z,
			Sections: [16]*chunkSection{},
		}

		loadedColumn := w.LoadColumn(x, z)
		if loadedColumn != nil {
			for _, section := range loadedColumn.Level.Sections {
				// anvil file contains this invalid, empty section for some reason: skip it
				if section.Y == 255 {
					continue
				}

				paletteLength := len(section.Palette)
				palette := make([]int32, paletteLength)
				for i, block := range section.Palette {
					palette[i] = data.ParseBlockId(block.Name, block.Properties)
				}
				// TODO: also create the palette map

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
					BlockStates: blockStates,
				}
			}
		}

		w.columnMap[x][z] = &val
		return &val
	}

	return w.columnMap[x][z]
}

// LoadColumn will load a column from a region file
func (w *World) LoadColumn(x int, z int) *anvilColumn {
	if w.anvilColumnMap == nil {
		w.anvilColumnMap = make(map[int]map[int]*anvilColumn)
	}

	if w.anvilColumnMap[x] == nil {
		w.anvilColumnMap[x] = make(map[int]*anvilColumn)
	}

	if w.anvilColumnMap[x][z] == nil {
		regionX := x >> 5
		regionZ := z >> 5
		r, err := region.Open(fmt.Sprintf("./test_world/region/r.%d.%d.mca", regionX, regionZ))
		if err != nil {
			return nil
		}
		defer r.Close()

		sectorX := x % 32
		if sectorX < 0 {
			sectorX += 32
		}
		sectorZ := z % 32
		if sectorZ < 0 {
			sectorZ += 32
		}

		sector, err := r.ReadSector(sectorZ, sectorX)
		if err != nil {
			return nil
		}

		var c anvilColumn
		err = c.Load(sector)
		if err != nil {
			return nil
		}

		w.anvilColumnMap[x][z] = &c
		return &c
	}

	return w.anvilColumnMap[x][z]
}
