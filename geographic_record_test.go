package geoindex

import (
    "fmt"
    "reflect"
    "testing"
    "time"
)

func TestGeographicRecord_Encode(t *testing.T) {
    now := time.Now()

    gr1 := NewGeographicRecord("source-name", "file.name1", now, true, 12.34, 34.56, nil)

    expected := map[string]interface{}{
        "timestamp":      now,
        "filepath":       "file.name1",
        "has_geographic": true,
        "latitude":       12.34,
        "longitude":      34.56,
        "s2_cell_id":     uint64(1610222620212933235),
        "source_name":    "source-name",
        "metadata":       make(map[string]interface{}),
        "comments":       make([]string, 0),
        "relationships":  make(map[string][]map[string]interface{}),
    }

    actual := gr1.Encode()

    if reflect.DeepEqual(actual, expected) != true {
        fmt.Printf("ACTUAL:\n")

        for k, v := range actual {
            fmt.Printf("[%s] = [%v]\n", k, v)
        }

        fmt.Printf("\n")

        fmt.Printf("EXPECTED:\n")

        for k, v := range expected {
            fmt.Printf("[%s] = [%v]\n", k, v)
        }

        fmt.Printf("\n")

        t.Fatalf("Encoding not correct.")
    }
}

func TestGeographicRecord_Metadata(t *testing.T) {
    metadata := map[string]interface{}{
        "aa": 123,
    }

    gr1 := NewGeographicRecord("source-name", "file.name1", time.Now(), true, 12.34, 34.56, metadata)

    if reflect.DeepEqual(metadata, gr1.Metadata) != true {
        t.Fatalf("Metadata not correct.")
    }
}

func TestGeographicRecord_AddComment(t *testing.T) {
    gr1 := NewGeographicRecord("source-name", "file.name1", time.Now(), true, 12.34, 34.56, nil)
    gr1.AddComment("comment1")

    expected := []string{
        "comment1",
    }

    actual := gr1.Comments()

    if reflect.DeepEqual(actual, expected) != true {
        t.Fatalf("Comments not correct.")
    }
}

func TestGeographicRecord_AddRelated(t *testing.T) {
    gr1 := NewGeographicRecord("source-name", "file.name1", time.Now(), true, 12.34, 34.56, nil)

    gr2 := NewGeographicRecord("source-name2", "file.name2", time.Now(), true, 12.34, 34.56, nil)
    gr1.AddRelated(gr2, "estranged")

    expected := map[string][]*GeographicRecord{
        "estranged": []*GeographicRecord{gr2},
    }

    actual := gr1.Relationships()

    if reflect.DeepEqual(actual, expected) != true {
        t.Fatalf("Relationships not correct.")
    }
}
