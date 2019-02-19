package geoindex

import (
	"bytes"
	"fmt"
	"path"
	"testing"
	"time"

	"github.com/dsoprea/go-logging"
)

func TestIndex_ExportGpx(t *testing.T) {
	index := NewTimeIndex()
	gc := NewGeographicCollector(index)

	err := RegisterImageFileProcessors(gc)
	log.PanicIf(err)

	err = RegisterDataFileProcessors(gc)
	log.PanicIf(err)

	err = gc.ReadFromPath(testAssetsPath)
	log.PanicIf(err)

	buffer := new(bytes.Buffer)

	err = gc.index.ExportGpx(buffer)
	log.PanicIf(err)

	expected := `<?xml version="1.0" encoding="UTF-8"?>
<gpx xmlns="http://www.topografix.com/GPX/1/1" xmlns:gpxx="http://www.garmin.com/xmlschemas/GpxExtensions/v3" gpxtpx="http://www.garmin.com/xmlschemas/TrackPointExtension/v1" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.topografix.com/GPX/1/1/gpx.xsd http://www.garmin.com/xmlschemas/GpxExtensions/v3 http://www.garmin.com/xmlschemas/GpxExtensionsv3.xsd http://www.garmin.com/xmlschemas/TrackPointExtension/v1 http://www.garmin.com/xmlschemas/TrackPointExtensionv1.xsd">
  <trk>
    <trkseg>
      <trkpt lat="47.644548" lon="-122.326897">
        <time>2009-10-17T18:37:26+0000</time>
      </trkpt>
      <trkpt lat="47.644548" lon="-122.326897">
        <time>2009-10-17T18:37:31+0000</time>
      </trkpt>
      <trkpt lat="47.644548" lon="-122.326897">
        <time>2009-10-17T18:37:34+0000</time>
      </trkpt>
      <trkpt lat="26.586666666666666" lon="-80.05361111111111">
        <time>2018-06-09T01:07:30+0000</time>
      </trkpt>
    </trkseg>
  </trk>
</gpx>`

	actual := buffer.String()

	if actual != expected {
		t.Fatalf("GPX data not correct:\n%s", actual)
	}
}

func ExampleIndex_ExportGpx() {
	index := NewTimeIndex()
	gc := NewGeographicCollector(index)

	err := RegisterImageFileProcessors(gc)
	log.PanicIf(err)

	err = RegisterDataFileProcessors(gc)
	log.PanicIf(err)

	err = gc.ReadFromPath(testAssetsPath)
	log.PanicIf(err)

	buffer := new(bytes.Buffer)

	err = gc.index.ExportGpx(buffer)
	log.PanicIf(err)

	output := buffer.String()
	fmt.Println(output)

	// Output:
	// <?xml version="1.0" encoding="UTF-8"?>
	// <gpx xmlns="http://www.topografix.com/GPX/1/1" xmlns:gpxx="http://www.garmin.com/xmlschemas/GpxExtensions/v3" gpxtpx="http://www.garmin.com/xmlschemas/TrackPointExtension/v1" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.topografix.com/GPX/1/1/gpx.xsd http://www.garmin.com/xmlschemas/GpxExtensions/v3 http://www.garmin.com/xmlschemas/GpxExtensionsv3.xsd http://www.garmin.com/xmlschemas/TrackPointExtension/v1 http://www.garmin.com/xmlschemas/TrackPointExtensionv1.xsd">
	//   <trk>
	//     <trkseg>
	//       <trkpt lat="47.644548" lon="-122.326897">
	//         <time>2009-10-17T18:37:26+0000</time>
	//       </trkpt>
	//       <trkpt lat="47.644548" lon="-122.326897">
	//         <time>2009-10-17T18:37:31+0000</time>
	//       </trkpt>
	//       <trkpt lat="47.644548" lon="-122.326897">
	//         <time>2009-10-17T18:37:34+0000</time>
	//       </trkpt>
	//       <trkpt lat="26.586666666666666" lon="-80.05361111111111">
	//         <time>2018-06-09T01:07:30+0000</time>
	//       </trkpt>
	//     </trkseg>
	//   </trk>
	// </gpx>
}

func ExampleIndex_Add() {
	index := NewTimeIndex()

	epochUtc := (time.Time{}).UTC()
	hasGeographic := true
	latitude := float64(123.456)
	longitude := float64(789.012)
	var metadata interface{}

	index.Add(SourceGeographicGpx, "data.gpx", epochUtc, hasGeographic, latitude, longitude, metadata)

	for _, te := range index.Series() {
		item := te.Items[0]
		gr := item.(GeographicRecord)

		timestampPhrase, err := te.Time.MarshalText()
		log.PanicIf(err)

		fmt.Printf("[%s] [%s] [%v] (%.10f) (%.10f)\n", string(timestampPhrase), path.Base(gr.Filepath), gr.HasGeographic, gr.Latitude, gr.Longitude)
	}

	// Output:
	// [0001-01-01T00:00:00Z] [data.gpx] [true] (123.4560000000) (789.0120000000)
}
