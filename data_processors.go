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

func GpxDataFileProcessor(index *Index, filepath string) (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    f, err := os.Open(filepath)
    log.PanicIf(err)

    defer f.Close()

    tpc := func(tp *gpxcommon.TrackPoint) (err error) {
        if tp.Time.IsZero() == true {
            dataLogger.Warningf(nil, "Skipping zero-time record: [%s] %s", filepath, tp)
            return nil
        }

        s2CellId := S2CellIdFromCoordinates(tp.LatitudeDecimal, tp.LongitudeDecimal)

        index.Add(
            filepath,
            tp.Time,
            tp.LatitudeDecimal,
            tp.LongitudeDecimal,
            s2CellId)

        return nil
    }

    err = gpxreader.EnumerateTrackPoints(f, tpc)
    log.PanicIf(err)

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
