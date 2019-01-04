package geoindex

import (
	"fmt"
	"io"
	"time"

	"github.com/dsoprea/go-gpx/writer"
	"github.com/dsoprea/go-logging"
	"github.com/dsoprea/go-time-index"
	"github.com/randomingenuity/go-utility/geographic"
)

type Index struct {
	ts timeindex.TimeSlice
}

func NewIndex() (gi *Index) {
	return &Index{
		ts: make(timeindex.TimeSlice, 0),
	}
}

func (index *Index) Series() timeindex.TimeSlice {
	return index.ts
}

// GeographicRecord describes a single bit of geographic information, whether
// it's a actual geographic data entry or an image with geographic data. Note
// that the naming is a bit of a misnomer since an image may not have
// geographic data and we might need to *derive* this from geographic data.
type GeographicRecord struct {
	Filepath      string
	HasGeographic bool
	Latitude      float64
	Longitude     float64
	S2CellId      uint64
	SourceName    string
	Metadata      interface{}
}

func (gr GeographicRecord) String() string {
	return fmt.Sprintf("GeographicRecord<F=[%s] LAT=[%.6f] LON=[%.6f] CELL=[%d]>", gr.Filepath, gr.Latitude, gr.Longitude, gr.S2CellId)
}

func (index *Index) Add(sourceName string, filepath string, timestamp time.Time, hasGeographic bool, latitude float64, longitude float64, metadata interface{}) {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	gr := GeographicRecord{
		SourceName:    sourceName,
		Filepath:      filepath,
		HasGeographic: hasGeographic,
		Metadata:      metadata,
	}

	if hasGeographic == true {
		gr.Latitude = latitude
		gr.Longitude = longitude

		gr.S2CellId = rigeo.S2CellIdFromCoordinates(latitude, longitude)
	}

	index.ts = index.ts.Add(timestamp, gr)

	// TODO(dustin): !! Convert our file-processors to an interface, implement a Name() method, and then store that name (or the processor) with the data.
	// TODO(dustin): !! Once we do the above, we can stop passing `sourceName` explicitly.

}

func (index *Index) ExportGpx(w io.Writer) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	b := gpxwriter.NewBuilder(w)
	gb := b.Gpx()

	tb, err := gb.Track()
	log.PanicIf(err)

	tsb, err := tb.TrackSegment()
	log.PanicIf(err)

	for _, timeItem := range index.ts {
		for _, item := range timeItem.Items {
			gr := item.(GeographicRecord)

			if gr.HasGeographic == true {
				tpb := tsb.TrackPoint()

				tpb.LatitudeDecimal = gr.Latitude
				tpb.LongitudeDecimal = gr.Longitude
				tpb.Time = timeItem.Time

				err = tpb.Write()
				log.PanicIf(err)
			}
		}
	}

	err = tsb.EndTrackSegment()
	log.PanicIf(err)

	err = tb.EndTrack()
	log.PanicIf(err)

	gb.EndGpx()

	return nil
}
