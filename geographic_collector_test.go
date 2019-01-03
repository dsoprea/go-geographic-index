package geoindex

import (
	"fmt"
	"reflect"
	"testing"
	"path"

	"github.com/dsoprea/go-logging"
)

func TestGeographicCollector_ReadFromPath_Images(t *testing.T) {
    index := NewIndex()
	gc := NewGeographicCollector(index)

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

func ExampleGeographicCollector_ReadFromPath() {
    index := NewIndex()
	gc := NewGeographicCollector(index)

	err := RegisterImageFileProcessors(gc)
	log.PanicIf(err)

	err = RegisterDataFileProcessors(gc)
	log.PanicIf(err)

	err = gc.ReadFromPath(testAssetsPath)
	log.PanicIf(err)

	for _, te := range index.Series() {
		item := te.Items[0]
		gr := item.(GeographicRecord)

		timestampPhrase, err := te.Time.MarshalText()
		log.PanicIf(err)

		fmt.Printf("[%s] [%s] [%v] (%.10f) (%.10f)\n", string(timestampPhrase), path.Base(gr.Filepath), gr.HasGeographic, gr.Latitude, gr.Longitude)
	}

	// Output:
	// [2009-10-17T18:37:26Z] [data.gpx] [true] (47.6445480000) (-122.3268970000)
	// [2009-10-17T18:37:31Z] [data.gpx] [true] (47.6445480000) (-122.3268970000)
	// [2009-10-17T18:37:34Z] [data.gpx] [true] (47.6445480000) (-122.3268970000)
	// [2018-06-09T01:07:30Z] [gps.jpg] [true] (26.5866666667) (-80.0536111111)
	// [2018-11-30T13:01:49Z] [IMG_20181130_1301493.jpg] [false] (0.0000000000) (0.0000000000)
}
