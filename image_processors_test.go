package geoindex

import (
	"path"
	"testing"

	"github.com/dsoprea/go-logging"
)

func TestJpegImageFileProcessor(t *testing.T) {
	index := NewTimeIndex()

	filepath := path.Join(testAssetsPath, "gps.jpg")

	jifp := NewJpegImageFileProcessor()

	err := jifp.Process(index, nil, filepath)
	log.PanicIf(err)

	if len(index.ts) != 1 {
		t.Fatalf("Exactly one index entry wasn't found: %v", index.ts)
	}

	actualFilepath := index.ts[0].Items[0].(*GeographicRecord).Filepath
	if actualFilepath != filepath {
		t.Fatalf("FIle-path of index entry is not correct: [%s]", actualFilepath)
	}
}

func TestRegisterImageFileProcessors(t *testing.T) {
	index := NewTimeIndex()
	gc := NewGeographicCollector(index, nil)

	err := RegisterImageFileProcessors(gc)
	log.PanicIf(err)
}
