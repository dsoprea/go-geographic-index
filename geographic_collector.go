package geoindex

import (
	"path"
	"strings"

	"github.com/dsoprea/go-logging"
	"github.com/randomingenuity/go-utility/filesystem"
)

var (
	imagesLogger = log.NewLogger("geoindex.images")
)

// TODO(dustin): !! Rename this file and tests to "collector.go". There's nothing geographic-specific about this file.

type GeographicCollector struct {
	processors        map[string]FileProcessor
	ti                *TimeIndex
	gi                *GeographicIndex
	filepathCollector []string
	visitedCount      int
}

type FileProcessor interface {
	Name() string
	Process(ti *TimeIndex, gi *GeographicIndex, filepath string) (err error)
}

// NewGeographicCollector takes both indices and populates them as files are
// processed. Either of them can be `nil` and, if that is the case, that index
// will not be utilized.
func NewGeographicCollector(ti *TimeIndex, gi *GeographicIndex) (gc *GeographicCollector) {
	filepathCollector := make([]string, 0)

	return &GeographicCollector{
		processors:        make(map[string]FileProcessor),
		ti:                ti,
		gi:                gi,
		filepathCollector: filepathCollector,
	}
}

func (gc *GeographicCollector) VisitedCount() int {
	return gc.visitedCount
}

// VisitedFilepaths returns the list of file-paths that we encountered.
func (gc *GeographicCollector) VisitedFilepaths() []string {
	return gc.filepathCollector
}

// AddFileProcessor registers a given processor for a given extension. The
// extension is case-insensitive and must include the initial period. This can
// be called more than once with one processor but not more than once for an
// extension.
func (gc *GeographicCollector) AddFileProcessor(extension string, processor FileProcessor) (err error) {
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

func (gc *GeographicCollector) ReadFromFilepath(filepath string) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	extension := path.Ext(filepath)
	extension = strings.ToLower(extension)

	fp, found := gc.processors[extension]

	// We don't have a processor for this type of file.
	if found == false {
		return nil
	}

	gc.filepathCollector = append(gc.filepathCollector, filepath)

	gc.visitedCount++

	err = fp.Process(gc.ti, gc.gi, filepath)
	log.PanicIf(err)

	return nil
}

func (gc *GeographicCollector) ReadFromPath(rootPath string) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	filesC, _, errC := rifs.ListFiles(rootPath, nil)

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
				err := gc.ReadFromFilepath(vf.Filepath)
				log.PanicIf(err)
			}
		}
	}

	return nil
}
