package game

import (
	"fmt"
	"github.com/Tnze/go-mc/save/region"
)

type World struct {
	columnMap map[int]map[int]*anvilColumn
}

func (w *World) GetChunk(x int, z int) (*anvilColumn, error) {
	if w.columnMap == nil {
		w.columnMap = make(map[int]map[int]*anvilColumn)
	}

	if w.columnMap[x] == nil {
		w.columnMap[x] = make(map[int]*anvilColumn)
	}

	if w.columnMap[x][z] == nil {
		regionX := x >> 5
		regionZ := z >> 5
		r, err := region.Open(fmt.Sprintf("./test_world/region/r.%d.%d.mca", regionX, regionZ))
		if err != nil {
			return nil, err
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

		data, err := r.ReadSector(sectorZ, sectorX)
		if err != nil {
			return nil, err
		}

		var c anvilColumn
		err = c.Load(data)
		if err != nil {
			return nil, err
		}

		w.columnMap[x][z] = &c
		return &c, nil
	}

	return w.columnMap[x][z], nil
}
