package geoindex

import (
	"testing"
	"time"

	"github.com/dsoprea/go-logging"
	"github.com/golang/geo/s2"
)

func TestGeographicIndex_AddWithRecord(t *testing.T) {
	gi := NewGeographicIndex()

	now := time.Now()
	gr := NewGeographicRecord("test source", "test file", now, true, 12.345678, 23.456789, nil)

	err := gi.AddWithRecord(gr)
	log.PanicIf(err)

	len_ := len(gi.s2Index)
	if len_ != 24 {
		t.Fatalf("Not enough entries in the index: (%d)", len_)
	}

	leafCellId := s2.CellID(gr.S2CellId)

	for parentCellIdRaw, indexed := range gi.s2Index {
		parentCellId := s2.CellID(parentCellIdRaw)
		if parentCellId.Contains(leafCellId) == false {
			t.Fatalf("indexed cell [%s] does not include leaf cell [%s].", parentCellId.ToToken(), leafCellId.ToToken())
		} else if len(indexed) != 1 {
			t.Fatalf("Indexed cell [%s] doesn't have exactly one record: (%d)", leafCellId.ToToken(), len(indexed))
		} else if indexed[0] != gr {
			t.Fatalf("Indexed cell [%s] doesn't have the right record: %s", leafCellId.ToToken(), *indexed[0])
		}
	}
}

func TestGeographicIndex_GetWithCoordinatesMetroLimited(t *testing.T) {
	gi := NewGeographicIndex()

	now := time.Now()
	gr := NewGeographicRecord("test source", "test file", now, true, 12.345678, 23.456789, nil)

	err := gi.AddWithRecord(gr)
	log.PanicIf(err)

	// Get an exact match.

	results, err := gi.GetWithCoordinatesMetroLimited(12.345678, 23.456789)
	log.PanicIf(err)

	if len(results) != 1 {
		t.Fatalf("Exactly one result was not found.")
	} else if results[0] != gr {
		t.Fatalf("Did not find the correct record: %v", results)
	}

	// Get an approximate match.

	results, err = gi.GetWithCoordinatesMetroLimited(12.3457, 23.4568)
	log.PanicIf(err)

	if len(results) != 1 {
		t.Fatalf("Exactly one result was not found.")
	} else if results[0] != gr {
		t.Fatalf("Did not find the correct record: %v", results)
	}
}

func TestGeographicIndex_GetWithCoordinates(t *testing.T) {
	gi := NewGeographicIndex()

	now := time.Now()
	gr := NewGeographicRecord("test source", "test file", now, true, 12.345678, 23.456789, nil)

	err := gi.AddWithRecord(gr)
	log.PanicIf(err)

	// Get an exact match.

	results, err := gi.GetWithCoordinates(12.345678, 23.456789, 30)
	log.PanicIf(err)

	if len(results) != 1 {
		t.Fatalf("Exactly one result was not found.")
	} else if results[0] != gr {
		t.Fatalf("Did not find the correct record: %v", results)
	}

	// Get an approximate match at as tight a resolution as possible.

	results, err = gi.GetWithCoordinates(12.345600, 23.456700, 14)
	log.PanicIf(err)

	if len(results) != 1 {
		t.Fatalf("Exactly one result was not found.")
	} else if results[0] != gr {
		t.Fatalf("Did not find the correct record: %v", results)
	}

	// Verify that we fail at any tighter of a resolution

	_, err = gi.GetWithCoordinates(12.345600, 23.456700, 15)
	if err != ErrNoNearMatch {
		t.Fatalf("Did not fail as expected when looking for a non-match: %v", err)
	}
}
