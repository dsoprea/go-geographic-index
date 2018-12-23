package geoindex

import (
// "github.com/dsoprea/go-time-index"
)

type Index struct {
}

func NewIndex() (gi *Index) {
	return new(Index)
}

func Add(filepath string, latitude float64, longitude float64) {

	// TODO(dustin): !! Finish. Calculate a Hilbert integer and store the coordinates and the integer in the index.

	// TODO(dustin): !! Convert our file-processors to an interface, implement a Name() method, and then store that name (or the processor) with the data.

}
