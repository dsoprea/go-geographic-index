package geoindex

import (
	"errors"

	"github.com/dsoprea/go-logging"
	"github.com/golang/geo/s2"
	"github.com/randomingenuity/go-utility/geographic"
)

const (
	// MinimumS2LevelForIndexing is the lowest level that we'll index an S2
	// coordinate at. This is the point at which we reach diminishing returns
	// on the space we use for the index.
	MinimumS2LevelForIndexing = 7
)

var (
	ErrNoGeographicInformation = errors.New("no geographic information")
	ErrNoNearMatch             = errors.New("no near match")
)

type GeographicIndex struct {
	s2Index map[uint64][]*GeographicRecord
}

func NewGeographicIndex() (gi *GeographicIndex) {
	return &GeographicIndex{
		s2Index: make(map[uint64][]*GeographicRecord),
	}
}

func (gi *GeographicIndex) AddWithRecord(gr *GeographicRecord) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	if gr.S2CellId == 0 {
		log.Panic(ErrNoGeographicInformation)
	}

	cellId := s2.CellID(gr.S2CellId)
	if cellId.IsLeaf() == false {
		log.Panicf("only leaf S2 cells are supported")
	}

	for level := cellId.Level(); level >= MinimumS2LevelForIndexing; level-- {
		parentCellId := cellId.Parent(level)
		parentCellIdRaw := uint64(parentCellId)

		if indexedTokens, found := gi.s2Index[parentCellIdRaw]; found == true {
			gi.s2Index[parentCellIdRaw] = append(indexedTokens, gr)
		} else {
			gi.s2Index[parentCellIdRaw] = []*GeographicRecord{gr}
		}
	}

	return nil
}

// GetWithCoordinatesMetroLimited will return anything between an exact match
// and the resolution of a general metropolitan area.
func (gi *GeographicIndex) GetWithCoordinatesMetroLimited(latitude, longitude float64) (results []*GeographicRecord, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	results, err = gi.GetWithCoordinates(latitude, longitude, MinimumS2LevelForIndexing)
	if err != nil {
		if err == ErrNoNearMatch {
			return nil, err
		}

		log.Panic(err)
	}

	return results, nil
}

// GetWithCoordinatesMetroLimited will return anything between an exact match
// and the resolution of a general metropolitan area. lowestAllowedLevel
// controls how large the area we're allowed to look for results in (the lower
// the larger).
func (gi *GeographicIndex) GetWithCoordinates(latitude, longitude float64, lowestAllowedLevel int) (results []*GeographicRecord, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	cellIdRaw := rigeo.S2CellFromCoordinates(latitude, longitude)
	cellId := s2.CellID(cellIdRaw)

	// Starting at finest resolution, iteratively search cells until we find
	// any indexed record.
	for level := cellId.Level(); level >= lowestAllowedLevel; level-- {
		parentCellId := cellId.Parent(level)
		parentCellIdRaw := uint64(parentCellId)

		indexedTokens := gi.s2Index[parentCellIdRaw]
		if indexedTokens == nil {
			continue
		}

		return indexedTokens, nil
	}

	return nil, ErrNoNearMatch
}
