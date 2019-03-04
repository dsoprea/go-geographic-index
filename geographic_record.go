package geoindex

import (
	"fmt"
	"path"
	"time"

	"github.com/randomingenuity/go-utility/geographic"
)

// GeographicRecord describes a single bit of geographic information, whether
// it's a actual geographic data entry or an image with geographic data. Note
// that the naming is a bit of a misnomer since an image may not have
// geographic data and we might need to *derive* this from geographic data.
type GeographicRecord struct {
	Timestamp     time.Time
	Filepath      string
	HasGeographic bool
	Latitude      float64
	Longitude     float64
	S2CellId      uint64
	SourceName    string
	Metadata      interface{}
}

func (gr GeographicRecord) String() string {
	return fmt.Sprintf("GeographicRecord<F=[%s] LAT=[%.6f] LON=[%.6f] CELL=[%d]>", path.Base(gr.Filepath), gr.Latitude, gr.Longitude, gr.S2CellId)
}

func NewGeographicRecord(sourceName string, filepath string, timestamp time.Time, hasGeographic bool, latitude float64, longitude float64, metadata interface{}) (gr *GeographicRecord) {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	gr = &GeographicRecord{
		SourceName:    sourceName,
		Timestamp:     timestamp,
		Filepath:      filepath,
		HasGeographic: hasGeographic,
		Metadata:      metadata,
	}

	if hasGeographic == true {
		gr.Latitude = latitude
		gr.Longitude = longitude

		cellIdRaw := rigeo.S2CellFromCoordinates(latitude, longitude)
		gr.S2CellId = uint64(cellIdRaw)
	}

	return gr
}
