package geoindex

import (
	"os"
	"path"
	"strings"

	"github.com/dsoprea/go-logging"
	"github.com/randomingenuity/go-utility/filesystem"
)

var (
	imagesLogger = log.NewLogger("geoindex.images")
)

// TODO(dustin): !! Rename this file and tests to "collector.go". There's nothing geographic-specific about this file.

// TODO(dustin): Add a mechanism to filter out aberrations in geographic data:
//
// Given a sliding window of (n) records in a geographic series sorted by
// timestamp, group by M number of right-side places on a hilbert position
// where they are adjacent in the series and close enough in time. Groups
// must be at least P members. For any two adjacent groups with at least one
// ungrouped record between them where these ungrouped members are also within
// a very close period of time to the rightmost member of the group on the left
// side and the leftmost member of the group on the right side log and ignore.
// This would help distill conflicing records that are present for some reason
// (maybe a data recording glitch or pictures from someone else) that would
// otherwise dramatically skew the grouping algorithm.

type GeographicCollector struct {
	processors map[string]FileProcessorFn
	ti         *TimeIndex
	gi         *GeographicIndex
}

// TODO(dustin): !! Convert to an interface and implement a Name() method.
type FileProcessorFn func(ti *TimeIndex, gi *GeographicIndex, filepath string) (err error)

// NewGeographicCollector takes both indices and populates them as files are
// processed. Either of them can be `nil` and, if that is the case, that index
// will not be utilized.
func NewGeographicCollector(ti *TimeIndex, gi *GeographicIndex) (gc *GeographicCollector) {
	return &GeographicCollector{
		processors: make(map[string]FileProcessorFn),
		ti:         ti,
		gi:         gi,
	}
}

// AddFileProcessor registers a given processor for a given extension. The
// extension is case-insensitive and must include the initial period. This can
// be called more than once with one processor but not more than once for an
// extension.
func (gc *GeographicCollector) AddFileProcessor(extension string, processor FileProcessorFn) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	extension = strings.ToLower(extension)

	_, found := gc.processors[extension]
	if found == true {
		log.Panicf("extension [%s] already registered", extension)
	}

	gc.processors[extension] = processor

	return nil
}

func (gc *GeographicCollector) ReadFromPath(rootPath string) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// Allow all directories and any file whose extension is associated with a
	// processor.
	filter := func(parent string, child os.FileInfo) (bool, error) {
		if child.IsDir() == true {
			return true, nil
		}

		extension := path.Ext(child.Name())
		extension = strings.ToLower(extension)

		_, found := gc.processors[extension]
		if found == true {
			return true, nil
		}

		return false, nil
	}

	filesC, errC := rifs.ListFiles(rootPath, filter)

FilesRead:

	for {
		select {
		case err, ok := <-errC:
			if ok == true {
				// TODO(dustin): Can we close these on the other side after sending and still get our data?
				close(filesC)
				close(errC)
			}

			log.PanicIf(err)

		case vf, ok := <-filesC:
			// We have finished reading. `vf` has an empty value.
			if ok == false {
				// The goroutine finished.
				break FilesRead
			}

			if vf.Info.IsDir() == false {
				extension := path.Ext(vf.Filepath)
				extension = strings.ToLower(extension)

				fp, found := gc.processors[extension]

				// Should never happen because the predicate above will already
				// ignore anything that doesn't have a processor.
				if found == false {
					log.Panicf("processor expected but not found (should never happen)")
				}

				err := fp(gc.ti, gc.gi, vf.Filepath)
				log.PanicIf(err)
			}
		}
	}

	return nil
}
