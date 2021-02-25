package util

import pk "gogs/impl/net/packet"

type CompactedDataArray struct {
	Data          []pk.Long
	bitsPerValue  int
	valuesPerLong int
	capacity      int
	bitMask       pk.Long
}

func (s *CompactedDataArray) Init(bitsPerValue int, capacity int) {
	s.bitsPerValue = bitsPerValue
	s.valuesPerLong = 64 / bitsPerValue
	s.Data = make([]pk.Long, capacity/s.valuesPerLong)
	s.capacity = capacity
	s.bitMask = (1 << bitsPerValue) - 1
}

func (s CompactedDataArray) Get(index int) pk.Long {
	dataIndex := index / s.valuesPerLong
	dataShift := (index % s.valuesPerLong) * s.bitsPerValue

	return (s.Data[dataIndex] >> dataShift) & s.bitMask

	/*
		// OLD FORMAT: values can span across longs
		index *= s.bitsPerValue
		dataIndex := index >> 6
		dataShift := index & 63


		val := s.Data[dataIndex] >> dataShift
		// check if value spreads over two longs
		if dataShift + s.bitsPerValue > 64 {
			val |= s.Data[dataIndex+1] << (64 - dataShift)
		}

		return val & s.bitMask
	*/
}

func (s *CompactedDataArray) Set(index int, val pk.Long) {
	dataIndex := index / s.valuesPerLong
	dataShift := (index % s.valuesPerLong) * s.bitsPerValue

	s.Data[dataIndex] &^= s.bitMask << dataShift
	s.Data[dataIndex] |= (val & s.bitMask) << dataShift

	/*
		// OLD FORMAT: values can span across longs
		index *= s.bitsPerValue
		dataIndex := index >> 6
		dataShift := index & 63

		// clear the bits needed to be set
		s.Data[dataIndex] &^= s.bitMask << dataShift
		s.Data[dataIndex] |= (val & s.bitMask) << dataShift

		// check if value spreads over two longs
		if dataShift + s.bitsPerValue > 64 {
			dataIndex += 1
			// clear the bits needed to be set
			s.Data[dataIndex] &^= pk.Long(1 << (dataShift + s.bitsPerValue - 64)) - 1
			s.Data[dataIndex] |= val >> (64 - dataShift)
		}
	*/
}
