package geoindex

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/dsoprea/go-logging"
)

func TestGeographicCollector_ReadFromPath_Images(t *testing.T) {
	gc := NewGeographicCollector()

	err := RegisterImageFileProcessors(gc)
	log.PanicIf(err)

	err = RegisterDataFileProcessors(gc)
	log.PanicIf(err)

	err = gc.ReadFromPath(testAssetsPath)
	log.PanicIf(err)

	actualTimestamps := make([]string, 0)

	for _, timeItem := range gc.index.ts {
		actualTimestamps = append(actualTimestamps, timeItem.Time.String())
	}

	expectedTimestamps := []string{
		"2009-10-17 18:37:26 +0000 UTC",
		"2009-10-17 18:37:31 +0000 UTC",
		"2009-10-17 18:37:34 +0000 UTC",
		"2018-06-09 01:07:30 +0000 UTC",
		"2018-11-30 13:01:49 +0000 UTC",
	}

	if reflect.DeepEqual(actualTimestamps, expectedTimestamps) != true {
		for i, timeItem := range gc.index.ts {
			fmt.Printf("(%d): %s\n", i, timeItem.Time)
		}

		t.Fatalf("Records incorrect (timestamps).")
	}

	actualData := make([]string, 0)

	for _, timeItem := range gc.index.ts {
		gr := timeItem.Items[0].(GeographicRecord)

		phrase := fmt.Sprintf("[%s] [%X]", timeItem.Time, gr.S2CellId)
		actualData = append(actualData, phrase)
	}

	actualData = actualData

	expectedData := []string{
		"[2009-10-17 18:37:26 +0000 UTC] [549014E3B65F8B85]",
		"[2009-10-17 18:37:31 +0000 UTC] [549014E3B65F8B85]",
		"[2009-10-17 18:37:34 +0000 UTC] [549014E3B65F8B85]",
		"[2018-06-09 01:07:30 +0000 UTC] [5ACC938D4BB4914B]",
		"[2018-11-30 13:01:49 +0000 UTC] [0]",
	}

	if reflect.DeepEqual(actualData, expectedData) != true {
		for i, timeItem := range gc.index.ts {
			gr := timeItem.Items[0].(GeographicRecord)

			fmt.Printf("(%d): [%s] [%X]\n", i, timeItem.Time, gr.S2CellId)
		}

		t.Fatalf("Records incorrect (data).")
	}
}
