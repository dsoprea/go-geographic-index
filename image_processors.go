package geoindex

import (
	"time"

	"io/ioutil"

	"github.com/dsoprea/go-exif"
	"github.com/dsoprea/go-jpeg-image-structure"
	"github.com/dsoprea/go-logging"
)

var (
	ipLogger = log.NewLogger("geoindex.image_processors")
)

type JpegImageFileProcessor struct {
	timestampSkew     time.Duration
	cameraModelFilter map[string]struct{}
}

func NewJpegImageFileProcessor() *JpegImageFileProcessor {
	return &JpegImageFileProcessor{
		cameraModelFilter: make(map[string]struct{}),
	}
}

func (jifp *JpegImageFileProcessor) SetImageTimestampSkew(timestampSkew time.Duration) {
	jifp.timestampSkew = timestampSkew
}

func (jifp *JpegImageFileProcessor) AddCameraModelToFilter(cameraModel string) {
	jifp.cameraModelFilter[cameraModel] = struct{}{}
}

func (jifp *JpegImageFileProcessor) Name() string {
	return "JpegImageFileProcessor"
}

func (jifp *JpegImageFileProcessor) getFirstExifTagStringValue(ifd *exif.Ifd, tagName string) (value string, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	results, err := ifd.FindTagWithName(tagName)
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

		valueRaw, err := ifd.TagValue(ite)
		log.PanicIf(err)

		value = valueRaw.(string)
	}

	return value, nil
}

func (jifp *JpegImageFileProcessor) Process(ti *TimeIndex, gi *GeographicIndex, filepath string) (err error) {
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
		if log.Is(err, exif.ErrNoExif) == true {
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
		gpsInfo, err := gpsIfd.GpsInfo()
		if err == nil {
			// Yes. We have geographic data.

			hasGeographicData = true
			latitude = gpsInfo.Latitude.Decimal()
			longitude = gpsInfo.Longitude.Decimal()
		} else {
			if log.Is(err, exif.ErrNoGpsTags) == false {
				log.Panic(err)
			}
		}
	}

	// Get the picture timestamp as stored in the EXIF.

	var timestamp time.Time

	exifIfd, err := rootIfd.ChildWithIfdPath(exif.IfdPathStandardExif)
	if err == nil {
		// We use this because it's the time that the image was captured versus the
		// time it was written to a file ("DateTimeDigitized"). Also, we have seen
		// "DateTime" being occasionally modified for a local timezone, but it's
		// rare and inconsistent. All things being equal, let's use something that
		// behaves consistently that we can account for.
		tagName := "DateTimeOriginal"

		timestampPhrase, err := jifp.getFirstExifTagStringValue(exifIfd, tagName)
		log.PanicIf(err)

		if timestampPhrase == "" {
			ipLogger.Warningf(nil, "Image has an empty timestamp: [%s]", filepath)
			return nil
		} else {
			timestamp, err = exif.ParseExifFullTimestamp(timestampPhrase)
			if err != nil {
				ipLogger.Warningf(nil, "Image's timestamp is unparseable: [%s] [%s]", filepath, timestampPhrase)
				return nil
			}

			timestamp = timestamp.Add(jifp.timestampSkew)
		}
	}

	// Get the camera model as stored in the EXIF. It will be empty here if
	// absent in the EXIF.

	// IFD-PATH=[IFD] ID=(0x0110) NAME=[Model] COUNT=(22) TYPE=[ASCII] VALUE=[Canon EOS 5D Mark III]
	tagName := "Model"

	cameraModel, err := jifp.getFirstExifTagStringValue(rootIfd, tagName)
	log.PanicIf(err)

	// Check the camera-model filter.
	if len(jifp.cameraModelFilter) > 0 {
		if _, found := jifp.cameraModelFilter[cameraModel]; found == false {
			return nil
		}
	}

	im := ImageMetadata{
		CameraModel: cameraModel,
	}

	gr := NewGeographicRecord(
		SourceImageJpeg,
		filepath,
		timestamp,
		hasGeographicData,
		latitude,
		longitude,
		im)

	if ti != nil {
		err := ti.AddWithRecord(gr)
		log.PanicIf(err)
	}

	if gi != nil {
		err := gi.AddWithRecord(gr)
		log.PanicIf(err)
	}

	return nil
}

// RegisterImageFileProcessors registers the processors for the image types
// that we know how to process. `imageTimestampSkew` is necessor to shift the
// timestamps of the EXIF times we read, which always appear as UTC.
func RegisterImageFileProcessors(gc *GeographicCollector, imageTimestampSkew time.Duration, cameraModels []string) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	jifp := NewJpegImageFileProcessor()
	jifp.SetImageTimestampSkew(imageTimestampSkew)

	if cameraModels != nil {
		for _, cameraModel := range cameraModels {
			jifp.AddCameraModelToFilter(cameraModel)
		}
	}

	err = gc.AddFileProcessor(".jpg", jifp)
	log.PanicIf(err)

	err = gc.AddFileProcessor(".jpeg", jifp)
	log.PanicIf(err)

	return nil
}
