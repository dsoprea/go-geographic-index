package geoindex

import (
	"testing"

	"github.com/dsoprea/go-logging"
)

func TestGeographicCollector_ReadFromPath_Images(t *testing.T) {
	gc := NewGeographicCollector()

	err := RegisterImageFileProcessors(gc)
	log.PanicIf(err)

	err = gc.ReadFromPath(testAssetsPath)
	log.PanicIf(err)

	if len(gc.index.ts) != 1 {
		t.Fatalf("Expected exactly one geographic record to be stored: %v\n", gc.index.ts)
	} else if gc.index.ts[0].Time.String() != "2018-04-29 01:22:57 +0000 UTC" {
		t.Fatalf("Timestamp of geographic record is not correct: [%s]", gc.index.ts[0].Time)
	}

	gr := gc.index.ts[0].Items[0].(GeographicRecord)
	if gr.S2CellId != 0x5ACC938D4BB4914B {
		t.Fatalf("S2 cell-ID of geographic record is not correct: (%0X)", gr.S2CellId)
	}
}
