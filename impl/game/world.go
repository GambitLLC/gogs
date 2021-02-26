package game

import (
	"github.com/Tnze/go-mc/save/region"
)

type World struct {
	columnMap map[int]map[int]*Column
}

func (w *World) GetChunk(x int, z int) (*Column, error) {
	if x < 0 || z < 0 {
		return nil, nil
	}

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
