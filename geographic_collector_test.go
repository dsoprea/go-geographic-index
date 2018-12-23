package geoindex

import (
	"testing"

	"github.com/dsoprea/go-logging"
)

func TestGeographicCollector_ReadFromPath(t *testing.T) {
	gc := NewGeographicCollector()

	err := RegisterImageFileProcessors(gc)
	log.PanicIf(err)

	err = gc.ReadFromPath(testAssetsPath)
	log.PanicIf(err)

	// TODO(dustin): !! Finish. We need to try and extract coordinates and store in the index. Then, we can verify.

	// if len(processed) != 1 {
	// 	t.Fatalf("Did not find the right number of files: %v\n", processed)
	// } else if processed[0] != path.Join(testAssetsPath, "IMG_20181130_1301493.jpg") {
	// 	t.Fatalf("Did not find the rightfile: %v\n", processed)
	// }
}
