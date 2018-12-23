package geoindex

import (
	"time"

	"github.com/dsoprea/go-time-index"
)

type Index struct {
	ts timeindex.TimeSlice
}

func NewIndex() (gi *Index) {
	return &Index{
		ts: make(timeindex.TimeSlice, 0),
	}
}

type GeographicRecord struct {
	Filepath  string
	Latitude  float64
	Longitude float64
	S2CellId  uint64
}

func (index *Index) Add(filepath string, timestamp time.Time, latitude float64, longitude float64, s2CellId uint64) {
	gr := GeographicRecord{
		Filepath:  filepath,
		Latitude:  latitude,
		Longitude: longitude,
		S2CellId:  s2CellId,
	}

	index.ts = index.ts.Add(timestamp, gr)

	// TODO(dustin): !! Convert our file-processors to an interface, implement a Name() method, and then store that name (or the processor) with the data.

}
