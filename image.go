package geoindex

import (
	"github.com/dsoprea/go-logging"
)

func JpegImageFileProcessor(index *Index, filepath string) (err error) {

	// TODO(dustin): !! Finish implementation.

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
