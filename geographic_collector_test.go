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
		"2018-04-29 01:22:57 +0000 UTC",
	}

	if reflect.DeepEqual(actualTimestamps, expectedTimestamps) != true {
		t.Fatalf("Records incorrect (timestamps): %v", actualTimestamps)
	}

	actualData := make([]string, 0)

	for _, timeItem := range gc.index.ts {
		gr := timeItem.Items[0].(GeographicRecord)

		phrase := fmt.Sprintf("[%s] [%X]", timeItem.Time, gr.S2CellId)
		actualData = append(actualData, phrase)
	}

	actualData = actualData

	expectedData := []string{
		"[2009-10-17 18:37:26 +0000 UTC] [1C58D0481A0B0C7B]",
		"[2009-10-17 18:37:31 +0000 UTC] [1C58D0481A0B0C7B]",
		"[2009-10-17 18:37:34 +0000 UTC] [1C58D0481A0B0C7B]",
		"[2018-04-29 01:22:57 +0000 UTC] [5ACC938D4BB4914B]",
	}

	if reflect.DeepEqual(actualData, expectedData) != true {
		t.Fatalf("Records incorrect (data): %v", actualData)
	}
}
