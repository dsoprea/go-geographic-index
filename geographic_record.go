package geoindex

import (
	"fmt"
	"path"
	"reflect"
	"time"

	"github.com/randomingenuity/go-utility/geographic"
)

type GeographicRecordRelatedTo struct {
	GeographicRecord *GeographicRecord
	Relationship     string
}

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
	comments      []string
	relatedTo     []GeographicRecordRelatedTo
}

func (gr GeographicRecord) String() string {
	return fmt.Sprintf("GeographicRecord<F=[%s] LAT=[%.6f] LON=[%.6f] CELL=[%d]>", path.Base(gr.Filepath), gr.Latitude, gr.Longitude, gr.S2CellId)
}

func (gr *GeographicRecord) Equal(other *GeographicRecord) bool {
	if other.Timestamp.Equal(gr.Timestamp) != true {
		return false
	} else if other.Filepath != gr.Filepath {
		return false
	} else if other.HasGeographic != gr.HasGeographic {
		return false
	} else if fmt.Sprintf("%.6f", other.Latitude) != fmt.Sprintf("%.6f", gr.Latitude) {
		return false
	} else if fmt.Sprintf("%.6f", other.Longitude) != fmt.Sprintf("%.6f", gr.Longitude) {
		return false
	} else if other.S2CellId != gr.S2CellId {
		return false
	} else if other.SourceName != gr.SourceName {
		return false
	} else if reflect.DeepEqual(other.Metadata, gr.Metadata) != true {
		return false
	} else if reflect.DeepEqual(other.comments, gr.comments) != true {
		return false
	}

	return true
}

func (gr *GeographicRecord) Dump() {
	fmt.Printf("GEOGRAPHIC RECORD\n")
	fmt.Printf("=================\n")

	fmt.Printf("Timestamp: [%s]\n", gr.Timestamp)
	fmt.Printf("Filepath: [%s]\n", gr.Filepath)
	fmt.Printf("HasGeographic: [%v]\n", gr.HasGeographic)
	fmt.Printf("Latitude: (%.6f)\n", gr.Latitude)
	fmt.Printf("Longitude: (%.6f)\n", gr.Longitude)
	fmt.Printf("S2CellId: (%d)\n", gr.S2CellId)
	fmt.Printf("SourceName: [%s]\n", gr.SourceName)
	fmt.Printf("Metadata: %v\n", gr.Metadata)

	fmt.Printf("\n")
}

func (gr *GeographicRecord) Encode() map[string]interface{} {
	relationships := gr.Relationships()
	encodedRelationships := make(map[string][]map[string]interface{})
	for type_, grList := range relationships {
		encodedGrList := make([]map[string]interface{}, len(grList))
		for i, gr := range grList {
			encodedGrList[i] = gr.Encode()
		}

		encodedRelationships[type_] = encodedGrList
	}

	return map[string]interface{}{
		"timestamp":      gr.Timestamp,
		"filepath":       gr.Filepath,
		"has_geographic": gr.HasGeographic,
		"latitude":       gr.Latitude,
		"longitude":      gr.Longitude,
		"s2_cell_id":     gr.S2CellId,
		"source_name":    gr.SourceName,
		"metadata":       gr.Metadata,
		"comments":       gr.Comments(),
		"relationships":  encodedRelationships,
	}
}

func (gr *GeographicRecord) Comments() []string {
	return gr.comments
}

func (gr *GeographicRecord) AddComment(comment string) {
	gr.comments = append(gr.comments, comment)
}

func (gr *GeographicRecord) Relationships() map[string][]*GeographicRecord {
	relationshipsMap := make(map[string][]*GeographicRecord)
	for _, grrt := range gr.relatedTo {
		if existing, found := relationshipsMap[grrt.Relationship]; found == true {
			existing = append(existing, grrt.GeographicRecord)
		} else {
			relationshipsMap[grrt.Relationship] = []*GeographicRecord{
				grrt.GeographicRecord,
			}
		}
	}

	return relationshipsMap
}

func (gr *GeographicRecord) AddRelated(relatedGr *GeographicRecord, relationship string) {
	grrt := GeographicRecordRelatedTo{
		GeographicRecord: relatedGr,
		Relationship:     relationship,
	}

	gr.relatedTo = append(gr.relatedTo, grrt)
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
		comments:      make([]string, 0),
		relatedTo:     make([]GeographicRecordRelatedTo, 0),
	}

	if hasGeographic == true {
		gr.Latitude = latitude
		gr.Longitude = longitude

		cellIdRaw := rigeo.S2CellFromCoordinates(latitude, longitude)
		gr.S2CellId = uint64(cellIdRaw)
	}

	return gr
}
