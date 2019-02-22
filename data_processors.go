package geoindex

import (
	"os"

	"github.com/dsoprea/go-gpx"
	"github.com/dsoprea/go-gpx/reader"
	"github.com/dsoprea/go-logging"
)

var (
	dataLogger = log.NewLogger("geoindex.data_processors")
)

func GpxDataFileProcessor(ti *TimeIndex, gi *GeographicIndex, filepath string) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	f, err := os.Open(filepath)
	log.PanicIf(err)

	defer f.Close()

	counter := 0

	tpc := func(tp *gpxcommon.TrackPoint) (err error) {
		if tp.Time.IsZero() == true {
			dataLogger.Warningf(nil, "Skipping zero-time record: [%s] %s", filepath, tp)
			return nil
		}

		gr := NewGeographicRecord(
			SourceGeographicGpx,
			filepath,
			tp.Time,
			true,
			tp.LatitudeDecimal,
			tp.LongitudeDecimal,
			nil)

		if ti != nil {
			err := ti.AddWithRecord(gr)
			log.PanicIf(err)
		}

		if gi != nil {
			err := gi.AddWithRecord(gr)
			log.PanicIf(err)
		}

		counter++

		return nil
	}

	err = gpxreader.EnumerateTrackPoints(f, tpc)
	log.PanicIf(err)

	dataLogger.Infof(nil, "Read (%d) records from [%s].", counter, filepath)

	return nil
}

func RegisterDataFileProcessors(gc *GeographicCollector) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	err = gc.AddFileProcessor(".gpx", GpxDataFileProcessor)
	log.PanicIf(err)

	return nil
}
