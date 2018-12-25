package geoindex

import (
	"io/ioutil"

	"github.com/dsoprea/go-exif"
	"github.com/dsoprea/go-jpeg-image-structure"
	"github.com/dsoprea/go-logging"
)

var (
	ipLogger = log.NewLogger("geoindex.image_processors")
)

func JpegImageFileProcessor(index *Index, filepath string) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	jmp := jpegstructure.NewJpegMediaParser()

	data, err := ioutil.ReadFile(filepath)
	log.PanicIf(err)

	sl, err := jmp.ParseBytes(data)
	log.PanicIf(err)

	rootIfd, _, err := sl.Exif()
	if err != nil {
		// Skip if it doesn't have EXIF data.
		if log.Is(err, jpegstructure.ErrNoExif) == true {
			return nil
		}

		log.Panic(err)
	}

	gpsIfd, err := rootIfd.ChildWithIfdPath(exif.IfdPathStandardGps)
	if err != nil {
		// Skip if no GPS data.
		if log.Is(err, exif.ErrTagNotFound) == true {
			return nil
		}

		log.Panic(err)
	}

	gi, err := gpsIfd.GpsInfo()
	if err != nil {
		ipLogger.Errorf(nil, err, "Could not extract GPS info: [%s]", filepath)
		return nil
	}

	index.Add(
		filepath,
		gi.Timestamp,
		gi.Latitude.Decimal(),
		gi.Longitude.Decimal(),
		uint64(gi.S2CellId()))

	return nil
}

func RegisterImageFileProcessors(gc *GeographicCollector) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	err = gc.AddFileProcessor(".jpg", JpegImageFileProcessor)
	log.PanicIf(err)

	err = gc.AddFileProcessor(".jpeg", JpegImageFileProcessor)
	log.PanicIf(err)

	return nil
}
