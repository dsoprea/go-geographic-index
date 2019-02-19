package geoindex

import (
	"io/ioutil"
	"time"

	"github.com/dsoprea/go-exif"
	"github.com/dsoprea/go-jpeg-image-structure"
	"github.com/dsoprea/go-logging"
)

var (
	ipLogger = log.NewLogger("geoindex.image_processors")
)

func getFirstExifTagStringValue(rootIfd *exif.Ifd, tagName string) (value string, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	results, err := rootIfd.FindTagWithName(tagName)
	if err != nil {
		if log.Is(err, exif.ErrTagNotFound) == true {
			results = nil
		} else {
			log.Panic(err)
		}
	} else {
		if len(results) == 0 {
			results = nil
		}
	}

	if results != nil {
		ite := results[0]

		valueRaw, err := rootIfd.TagValue(ite)
		log.PanicIf(err)

		value = valueRaw.(string)
	}

	return value, nil
}

func JpegImageFileProcessor(index *TimeIndex, filepath string) (err error) {
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

	hasGps := true
	if err != nil {
		// Skip if no GPS data.
		if log.Is(err, exif.ErrTagNotFound) == true {
			hasGps = false
		} else {
			log.Panic(err)
		}
	}

	hasGeographicData := false
	var latitude float64
	var longitude float64

	if hasGps == true {
		gi, err := gpsIfd.GpsInfo()

		if err == nil {
			// Yes. We have geographic data.

			hasGeographicData = true
			latitude = gi.Latitude.Decimal()
			longitude = gi.Longitude.Decimal()
		}
	}

	// Get the picture timestamp as stored in the EXIF.

	tagName := "DateTime"

	timestampPhrase, err := getFirstExifTagStringValue(rootIfd, tagName)
	log.PanicIf(err)

	var timestamp time.Time
	if timestampPhrase == "" {
		ipLogger.Warningf(nil, "Image has an empty timestamp: [%s]", filepath)
		return nil
	} else {
		timestamp, err = exif.ParseExifFullTimestamp(timestampPhrase)
		if err != nil {
			ipLogger.Warningf(nil, "Image's timestamp is unparseable: [%s] [%s]", filepath, timestampPhrase)
			return nil
		}
	}

	// Get the camera model as stored in the EXIF. It will be empty here if
	// absent in the EXIF.

	// IFD-PATH=[IFD] ID=(0x0110) NAME=[Model] COUNT=(22) TYPE=[ASCII] VALUE=[Canon EOS 5D Mark III]
	tagName = "Model"

	cameraModel, err := getFirstExifTagStringValue(rootIfd, tagName)
	log.PanicIf(err)

	im := ImageMetadata{
		CameraModel: cameraModel,
	}

	index.Add(
		SourceImageJpeg,
		filepath,
		timestamp,
		hasGeographicData,
		latitude,
		longitude,
		im)

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
