package geoindex

import (
	"fmt"
	"io"
	"time"

	"github.com/dsoprea/go-gpx/writer"
	"github.com/dsoprea/go-logging"
	"github.com/dsoprea/go-time-index"
)

type Index struct {
	ts timeindex.TimeSlice
}

func NewIndex() (gi *Index) {
	return &Index{
		ts: make(timeindex.TimeSlice, 0),
	}
}

type GeographicRecord struct {
	Filepath  string
	Latitude  float64
	Longitude float64
	S2CellId  uint64
}

func (gr GeographicRecord) String() string {
	return fmt.Sprintf("GeographicRecord<F=[%s] LAT=[%.6f] LON=[%.6f] CELL=[%d]>", gr.Filepath, gr.Latitude, gr.Longitude, gr.S2CellId)
}

func (index *Index) Add(filepath string, timestamp time.Time, latitude float64, longitude float64, s2CellId uint64) {
	gr := GeographicRecord{
		Filepath:  filepath,
		Latitude:  latitude,
		Longitude: longitude,
		S2CellId:  s2CellId,
	}

	index.ts = index.ts.Add(timestamp, gr)

	// TODO(dustin): !! Convert our file-processors to an interface, implement a Name() method, and then store that name (or the processor) with the data.

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

			tpb := tsb.TrackPoint()

			tpb.LatitudeDecimal = gr.Latitude
			tpb.LongitudeDecimal = gr.Longitude
			tpb.Time = timeItem.Time

			err = tpb.Write()
			log.PanicIf(err)
		}
	}

	err = tsb.EndTrackSegment()
	log.PanicIf(err)

	err = tb.EndTrack()
	log.PanicIf(err)

	gb.EndGpx()

	return nil
}
