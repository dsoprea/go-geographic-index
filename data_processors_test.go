package geoindex

import (
	"path"
	"reflect"
	"testing"

	"github.com/dsoprea/go-logging"
)

func TestGpxDataFileProcessor_SkipNoTimeRecords(t *testing.T) {
	index := NewTimeIndex()

	filepath := path.Join(testAssetsPath, "no_times.gpx")

	gdfp := NewGpxDataFileProcessor()

	err := gdfp.Process(index, nil, filepath)
	log.PanicIf(err)

	if len(index.ts) > 0 {
		t.Fatalf("Expected no records to be returned.")
	}
}

func TestGpxDataFileProcessor_WithTimeRecords(t *testing.T) {
	index := NewTimeIndex()

	filepath := path.Join(testAssetsPath, "data.gpx")

	gdfp := NewGpxDataFileProcessor()

	err := gdfp.Process(index, nil, filepath)
	log.PanicIf(err)

	actual := make([]string, 0)

	for _, timeItem := range index.ts {
		actual = append(actual, timeItem.Time.String())
	}

	expected := []string{
		"2009-10-17 18:37:26 +0000 UTC",
		"2009-10-17 18:37:31 +0000 UTC",
		"2009-10-17 18:37:34 +0000 UTC",
	}

	if reflect.DeepEqual(actual, expected) != true {
		t.Fatalf("Records incorrect: %v", actual)
	}
}
