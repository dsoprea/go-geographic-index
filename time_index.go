package geoindex

import (
	"io"
	"time"

	"github.com/dsoprea/go-gpx/writer"
	"github.com/dsoprea/go-logging"
	"github.com/dsoprea/go-time-index"
)

type TimeIndex struct {
	ts timeindex.TimeSlice
}

func NewTimeIndex() (ti *TimeIndex) {
	return &TimeIndex{
		ts: make(timeindex.TimeSlice, 0),
	}
}

func (index *TimeIndex) Series() timeindex.TimeSlice {
	return index.ts
}

func (index *TimeIndex) AddWithRecord(gr *GeographicRecord) (err error) {
	index.ts = index.ts.Add(gr.Timestamp, gr)

	return nil
}

// Add adds a record to the time-index. This method is obsolete. Please use
// `AddWithRecord` instead.
func (index *TimeIndex) Add(sourceName string, filepath string, timestamp time.Time, hasGeographic bool, latitude float64, longitude float64, metadata interface{}) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	gr := NewGeographicRecord(sourceName, filepath, timestamp, hasGeographic, latitude, longitude, metadata)

	err = index.AddWithRecord(gr)
	log.PanicIf(err)

	// TODO(dustin): !! Convert our file-processors to an interface, implement a Name() method, and then store that name (or the processor) with the data.
	// TODO(dustin): !! Once we do the above, we can stop passing `sourceName` explicitly.

	return nil
}

func (index *TimeIndex) ExportGpx(w io.Writer) (err error) {
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
			gr := item.(*GeographicRecord)

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
