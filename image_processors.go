package geoindex

import (
	"fmt"
	"os"
	"path"
	"time"

	"crypto/sha1"
	"encoding/gob"
	"io/ioutil"

	"github.com/dsoprea/go-exif"
	"github.com/dsoprea/go-jpeg-image-structure"
	"github.com/dsoprea/go-logging"
)

var (
	ipLogger = log.NewLogger("geoindex.image_processors")
)

var (
	imageCacheRootPath = ""
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

// Process extracts metadata from a single image.
func (jifp *JpegImageFileProcessor) Process(ti *TimeIndex, gi *GeographicIndex, filepath string) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	gr := new(GeographicRecord)

	cacheFilepath := ""
	if imageCacheRootPath != "" {
		// Calculate the SHA1.

		// Produce a hash of the original image location.
		h := sha1.New()

		_, err := h.Write([]byte(filepath))
		log.PanicIf(err)

		filepathSha1 := h.Sum(nil)

		filepathSha1Phrase := fmt.Sprintf("%x", filepathSha1)

		// Use the hash to make sure our filename is unique in the cache.
		cacheFilename := fmt.Sprintf("%s.%s", path.Base(filepath), filepathSha1Phrase)
		cacheFilepath = path.Join(imageCacheRootPath, cacheFilename)
	}

	// TODO(dustin): !! Why do we seem to crawl and then jump to the end when reading through the cache.

	c, err := os.Open(cacheFilepath)
	if err == nil {
		defer c.Close()

		// Load items from cache.

		gd := gob.NewDecoder(c)
		err := gd.Decode(gr)
		log.PanicIf(err)
	} else if os.IsNotExist(err) == false {
		// There's an issue reading. Try to remove.
		err := os.Remove(cacheFilepath)
		if err != nil {
			ipLogger.Errorf(nil, "Could not read or delete cache for image [%s]: [%s]", filepath, cacheFilepath)
			log.Panic(err)
		}
	} else {
		// Build items fresh.

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

		// Get the picture timestamp as stored in the EXIF. We try several
		// different tags.

		var timestamp time.Time

		exifIfd, err := rootIfd.ChildWithIfdPath(exif.IfdPathStandardExif)
		if err != nil && log.Is(err, exif.ErrTagNotFound) == false {
			log.Panic(err)
		}

		if err == nil {
			// We prefer this because it's the time that the image was captured
			// versus the time it was written to a file ("DateTimeDigitized"). Also,
			// we have seen "DateTime" being occasionally modified for a local
			// timezone, but it's rare and inconsistent. All things being equal,
			// let's use something that behaves consistently that we can account
			// for.
			tagName := "DateTimeOriginal"

			timestampPhrase, err := jifp.getFirstExifTagStringValue(exifIfd, tagName)
			log.PanicIf(err)

			if timestampPhrase == "" {
				ipLogger.Warningf(nil, "Image has an empty timestamp for [%s]: [%s]", tagName, filepath)
			} else {
				timestamp, err = exif.ParseExifFullTimestamp(timestampPhrase)
				if err != nil {
					ipLogger.Warningf(nil, "Image's [%s] timestamp is unparseable: [%s] [%s]", tagName, filepath, timestampPhrase)
				} else {
					timestamp = timestamp.Add(jifp.timestampSkew)
				}
			}

			if timestamp.IsZero() == true {
				tagName := "DateTimeDigitized"

				timestampPhrase, err := jifp.getFirstExifTagStringValue(exifIfd, tagName)
				log.PanicIf(err)

				if timestampPhrase == "" {
					ipLogger.Warningf(nil, "Image has an empty timestamp for [%s]: [%s]", tagName, filepath)
				} else {
					timestamp, err = exif.ParseExifFullTimestamp(timestampPhrase)
					if err != nil {
						ipLogger.Warningf(nil, "Image's [%s] timestamp is unparseable: [%s] [%s]", tagName, filepath, timestampPhrase)
					} else {
						timestamp = timestamp.Add(jifp.timestampSkew)
					}
				}
			}
		}

		if timestamp.IsZero() == true {
			tagName := "DateTime"

			timestampPhrase, err := jifp.getFirstExifTagStringValue(rootIfd, tagName)
			log.PanicIf(err)

			if timestampPhrase == "" {
				ipLogger.Warningf(nil, "Image has an empty timestamp for [%s]: [%s]", tagName, filepath)
			} else {
				timestamp, err = exif.ParseExifFullTimestamp(timestampPhrase)
				if err != nil {
					ipLogger.Warningf(nil, "Image's [%s] timestamp is unparseable: [%s] [%s]", tagName, filepath, timestampPhrase)
				} else {
					timestamp = timestamp.Add(jifp.timestampSkew)
				}
			}
		}

		if timestamp.IsZero() == true {
			ipLogger.Warningf(nil, "Image does not have a timestamp: [%s]", filepath)
			return nil
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

		gr = NewGeographicRecord(
			SourceImageJpeg,
			filepath,
			timestamp,
			hasGeographicData,
			latitude,
			longitude,
			im)

		if cacheFilepath != "" {
			g, err := os.OpenFile(cacheFilepath, os.O_CREATE|os.O_RDWR, 0644)
			log.PanicIf(err)

			defer g.Close()

			e := gob.NewEncoder(g)

			err = e.Encode(gr)
			log.PanicIf(err)
		}
	}

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

func init() {
	imageCacheRootPath = os.Getenv("GEOINDEX_IMAGE_INFO_CACHE_PATH")

	pathF, err := os.Open(imageCacheRootPath)
	if err != nil {
		if os.IsNotExist(err) == false {
			log.Panic(err)
		}

		os.MkdirAll(imageCacheRootPath, 0755)
	} else {
		pathF.Close()
	}

	gob.Register(&GeographicRecord{})
	gob.Register(ImageMetadata{})
}
