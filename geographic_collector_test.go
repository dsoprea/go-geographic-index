package geoindex

import (
	"errors"
	"fmt"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/dsoprea/go-logging"
)

func TestGeographicCollector_ReadFromPath_Images(t *testing.T) {
	index := NewTimeIndex()
	gc := NewGeographicCollector(index, nil)

	err := RegisterImageFileProcessors(gc)
	log.PanicIf(err)

	err = RegisterDataFileProcessors(gc)
	log.PanicIf(err)

	err = gc.ReadFromPath(testAssetsPath)
	log.PanicIf(err)

	actualTimestamps := make([]string, 0)

	for _, timeItem := range index.ts {
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
		for i, timeItem := range index.ts {
			fmt.Printf("(%d): %s\n", i, timeItem.Time)
		}

		t.Fatalf("Records incorrect (timestamps).")
	}

	actualData := make([]string, 0)

	for _, timeItem := range index.ts {
		gr := timeItem.Items[0].(*GeographicRecord)

		phrase := fmt.Sprintf("[%s] [%X]", timeItem.Time, gr.S2CellId)
		actualData = append(actualData, phrase)
	}

	actualData = actualData

	expectedData := []string{
		"[2009-10-17 18:37:26 +0000 UTC] [549014E3B65F8B85]",
		"[2009-10-17 18:37:31 +0000 UTC] [549014E3B65F8B85]",
		"[2009-10-17 18:37:34 +0000 UTC] [549014E3B65F8B85]",
		"[2018-06-09 01:07:30 +0000 UTC] [88D8D8FFD5C4DEAD]",
		"[2018-11-30 13:01:49 +0000 UTC] [0]",
	}

	if reflect.DeepEqual(actualData, expectedData) != true {
		for i, timeItem := range index.ts {
			gr := timeItem.Items[0].(*GeographicRecord)

			fmt.Printf("(%d): [%s] [%X]\n", i, timeItem.Time, gr.S2CellId)
		}

		t.Fatalf("Records incorrect (data).")
	}
}

func TestGeographicCollector_ReadFromPath_WithTimeAndGeographicIndices(t *testing.T) {
	ti := NewTimeIndex()
	gi := NewGeographicIndex()
	gc := NewGeographicCollector(ti, gi)

	err := RegisterDataFileProcessors(gc)
	log.PanicIf(err)

	err = gc.ReadFromPath(testAssetsPath)
	log.PanicIf(err)

	ts := ti.Series()

	d := time.Date(2009, 10, 17, 18, 37, 26, 0, time.UTC)
	i := ts.Search(d)

	if ts[i].Time != d {
		t.Fatalf("Could not find entry for time.")
	} else if len(ts[i].Items) != 1 {
		t.Fatalf("Exactly one record wasn't stored: %v", ts[i].Items)
	}

	recoveredTimestamp := ts[i].Items[0].(*GeographicRecord).Timestamp
	if recoveredTimestamp != d {
		t.Fatalf("Result did not have the right timestamp: %v", recoveredTimestamp)
	}

	results, err := gi.GetWithCoordinatesMetroLimited(47.6445480000, -122.3268970000)
	log.PanicIf(err)

	gr := results[0]
	if gr.Filepath != path.Join(testAssetsPath, "data.gpx") {
		t.Fatalf("First entry not correct (file): [%s]", gr.Filepath)
	} else if gr.Timestamp != time.Date(2009, 10, 17, 18, 37, 26, 0, time.UTC) {
		t.Fatalf("First entry not correct (time): [%v]", gr.Timestamp)
	}

	gr = results[1]
	if gr.Filepath != path.Join(testAssetsPath, "data.gpx") {
		t.Fatalf("First entry not correct (file): [%s]", gr.Filepath)
	} else if gr.Timestamp != time.Date(2009, 10, 17, 18, 37, 31, 0, time.UTC) {
		t.Fatalf("First entry not correct (time): [%v]", gr.Timestamp)
	}

	gr = results[2]
	if gr.Filepath != path.Join(testAssetsPath, "data.gpx") {
		t.Fatalf("First entry not correct (file): [%s]", gr.Filepath)
	} else if gr.Timestamp != time.Date(2009, 10, 17, 18, 37, 34, 0, time.UTC) {
		t.Fatalf("First entry not correct (time): [%v]", gr.Timestamp)
	}
}

func ExampleGeographicCollector_ReadFromPath() {
	index := NewTimeIndex()
	gc := NewGeographicCollector(index, nil)

	err := RegisterImageFileProcessors(gc)
	log.PanicIf(err)

	err = RegisterDataFileProcessors(gc)
	log.PanicIf(err)

	err = gc.ReadFromPath(testAssetsPath)
	log.PanicIf(err)

	for _, te := range index.Series() {
		item := te.Items[0]
		gr := item.(*GeographicRecord)

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

func ExampleGeographicCollector_ReadFromPath_WithTimeAndGeographicIndices() {
	ti := NewTimeIndex()
	gi := NewGeographicIndex()

	gc := NewGeographicCollector(ti, gi)

	err := RegisterDataFileProcessors(gc)
	log.PanicIf(err)

	err = gc.ReadFromPath(testAssetsPath)
	log.PanicIf(err)

	d := time.Date(2009, 10, 17, 18, 37, 26, 0, time.UTC)
	ts := ti.Series()
	i := ts.Search(d)

	if i >= len(ts) {
		panic(errors.New("Record not found for time (1)."))
	}

	result := ts[i]
	if result.Time != d {
		panic(errors.New("Record not found for time (2)."))
	}

	fmt.Printf("Search by time:\n")
	fmt.Printf("\n")

	for _, o := range result.Items {
		gr := o.(*GeographicRecord)
		fmt.Printf("%v\n", gr)
	}

	fmt.Printf("\n")
	fmt.Printf("Search by coordinate:\n")
	fmt.Printf("\n")

	results, err := gi.GetWithCoordinatesMetroLimited(47.6445480000, -122.3268970000)
	log.PanicIf(err)

	for _, gr := range results {
		fmt.Printf("%s (%.6f, %.6f)\n", gr.Timestamp, gr.Latitude, gr.Longitude)
	}

	// Output:
	// Search by time:
	//
	// GeographicRecord<F=[data.gpx] LAT=[47.644548] LON=[-122.326897] CELL=[6093393264082127749]>
	//
	// Search by coordinate:
	//
	// 2009-10-17 18:37:26 +0000 UTC (47.644548, -122.326897)
	// 2009-10-17 18:37:31 +0000 UTC (47.644548, -122.326897)
	// 2009-10-17 18:37:34 +0000 UTC (47.644548, -122.326897)
}
