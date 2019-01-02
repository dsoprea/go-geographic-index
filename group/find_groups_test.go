package geogroup

import (
    "testing"
    "time"
    "fmt"
    "path"
    "os"

    "github.com/dsoprea/go-geographic-index"
    "github.com/dsoprea/go-logging"
    "github.com/dsoprea/go-time-index"
    "github.com/dsoprea/go-geographic-attractor/index"
    "github.com/dsoprea/go-geographic-attractor/parse"
)

const (
    oneDay = time.Hour * 24
)

var (
    epochUtc = time.Unix(0, 0).UTC()
)

func TestFindGroups_AddUnassigned(t *testing.T) {
    fg := NewFindGroups(nil, nil, nil)

    gr := geoindex.GeographicRecord{
        S2CellId: 123,
    }

    reason := "some reason"

    fg.addUnassigned(gr, reason)

    unassignedRecords := fg.UnassignedRecords()

    if len(unassignedRecords) != 1 {
        t.Fatalf("There wasn't exactly one unassigned record: (%d)", len(fg.unassignedRecords))
    }

    ur := unassignedRecords[0]
    if ur.Geographic != gr {
        t.Fatalf("Geographic record not stored correctly.")
    } else if ur.Reason != reason {
        t.Fatalf("Reason not stored correctly.")
    }
}

func getTestLocationIndex(timeBase time.Time) (locationIndex *geoindex.Index) {
    locationIndex = geoindex.NewIndex()

    timeSeries := map[string]struct {
        timestamp time.Time
        latitude float64
        longitude float64
    } {
        "file00.gpx": { timeBase.Add(time.Hour * 0 + time.Minute * 0), 1.1, 10.1 },
        "file01.gpx": { timeBase.Add(time.Hour * 0 + time.Minute * 1), 1.2, 10.2 },
        "file02.gpx": { timeBase.Add(time.Hour * 0 + time.Minute * 2), 1.3, 10.3 },
        "file03.gpx": { timeBase.Add(time.Hour * 0 + time.Minute * 3), 1.4, 10.4 },
        "file04.gpx": { timeBase.Add(time.Hour * 0 + time.Minute * 4), 1.5, 10.5 },

        "file10.gpx": { timeBase.Add(time.Hour * 1 + time.Minute * 0), 2.1, 20.1 },
        "file11.gpx": { timeBase.Add(time.Hour * 1 + time.Minute * 5), 2.2, 20.2 },
        "file12.gpx": { timeBase.Add(time.Hour * 1 + time.Minute * 10), 2.3, 20.3 },
        "file13.gpx": { timeBase.Add(time.Hour * 1 + time.Minute * 15), 2.4, 20.4 },
        "file14.gpx": { timeBase.Add(time.Hour * 1 + time.Minute * 20), 2.5, 20.5 },

        "file20.gpx": { timeBase.Add(time.Hour * 2 + time.Minute * 0), 3.1, 30.1 },
        "file21.gpx": { timeBase.Add(time.Hour * 2 + time.Minute * 1), 3.2, 30.2 },
        "file22.gpx": { timeBase.Add(time.Hour * 2 + time.Minute * 2), 3.3, 30.3 },
        "file23.gpx": { timeBase.Add(time.Hour * 2 + time.Minute * 3), 3.4, 30.4 },
        "file24.gpx": { timeBase.Add(time.Hour * 2 + time.Minute * 4), 3.5, 30.5 },

        "file30.gpx": { timeBase.Add(time.Hour * 3 + time.Minute * 0), 4.1, 40.1 },
        "file31.gpx": { timeBase.Add(time.Hour * 3 + time.Minute * 10), 4.2, 40.2 },
        "file32.gpx": { timeBase.Add(time.Hour * 3 + time.Minute * 20), 4.3, 40.3 },
        "file33.gpx": { timeBase.Add(time.Hour * 3 + time.Minute * 30), 4.4, 40.4 },
        "file34.gpx": { timeBase.Add(time.Hour * 3 + time.Minute * 40), 4.5, 40.5 },

        "file40.gpx": { timeBase.Add(oneDay * 2 + time.Hour * 0 + time.Minute * 0), 5.1, 50.1 },
        "file41.gpx": { timeBase.Add(oneDay * 2 + time.Hour * 0 + time.Minute * 10), 5.2, 50.2 },
        "file42.gpx": { timeBase.Add(oneDay * 2 + time.Hour * 0 + time.Minute * 20), 5.3, 50.3 },
        "file43.gpx": { timeBase.Add(oneDay * 2 + time.Hour * 0 + time.Minute * 30), 5.4, 50.4 },
        "file44.gpx": { timeBase.Add(oneDay * 2 + time.Hour * 0 + time.Minute * 40), 5.5, 50.5 },

        "file50.gpx": { timeBase.Add(oneDay * 6 + time.Hour * 0 + time.Minute * 0), 6.1, 60.1 },
        "file51.gpx": { timeBase.Add(oneDay * 6 + time.Hour * 0 + time.Minute * 10), 6.2, 60.2 },
        "file52.gpx": { timeBase.Add(oneDay * 6 + time.Hour * 0 + time.Minute * 20), 6.3, 60.3 },
        "file53.gpx": { timeBase.Add(oneDay * 6 + time.Hour * 0 + time.Minute * 30), 6.4, 60.4 },
        "file54.gpx": { timeBase.Add(oneDay * 6 + time.Hour * 0 + time.Minute * 40), 6.5, 60.5 },
    }

    for filepath, x := range timeSeries {
        locationIndex.Add(geoindex.SourceGeographicGpx, filepath, x.timestamp, true, x.latitude, x.longitude, 0, nil)
    }

    return locationIndex
}

func TestFindGroups_FindLocationByTime_ExactMatch(t *testing.T) {
    timeBase := epochUtc
    locationIndex := getTestLocationIndex(timeBase)

    fg := NewFindGroups(locationIndex, nil, nil)
    
    imageTimestamp := timeBase.Add(time.Hour * 1 + time.Minute * 10)
    
    imageTe := timeindex.TimeEntry{
        Time: imageTimestamp,
        Items: nil,
    }

    matchedTe, err := fg.findLocationByTime(imageTe)
    log.PanicIf(err)

    expectedLocationTimestamp := timeBase.Add(time.Hour * 1 + time.Minute * 10)

    if matchedTe.Time != expectedLocationTimestamp {
        t.Fatalf("The matched location timestamp is not correct: [%s] != [%s]", matchedTe.Time, expectedLocationTimestamp)
    } else if len(matchedTe.Items) != 1 {
        t.Fatalf("Expected exactly one location item to be matched: %v\n", matchedTe.Items)
    }

    gr := matchedTe.Items[0].(geoindex.GeographicRecord)

    expectedLatitude := float64(2.3)
    if gr.Latitude != expectedLatitude {
        t.Fatalf("Matched latitude not correct: [%.10f] != [%.10f]", gr.Latitude, expectedLatitude)
    }

    expectedLongitude := float64(20.3)
    if gr.Longitude != expectedLongitude {
        t.Fatalf("Matched longitude not correct: [%.10f] != [%.10f]", gr.Longitude, expectedLongitude)
    }
}

func TestFindGroups_FindLocationByTime_JustBeforeLocationRecord(t *testing.T) {
    timeBase := epochUtc
    locationIndex := getTestLocationIndex(timeBase)

    fg := NewFindGroups(locationIndex, nil, nil)
    
    imageTimestamp := timeBase.Add(time.Hour * 1 + time.Minute * 9)
    
    imageTe := timeindex.TimeEntry{
        Time: imageTimestamp,
        Items: nil,
    }

    matchedTe, err := fg.findLocationByTime(imageTe)
    log.PanicIf(err)

    expectedLocationTimestamp := timeBase.Add(time.Hour * 1 + time.Minute * 10)

    if matchedTe.Time != expectedLocationTimestamp {
        t.Fatalf("The matched location timestamp is not correct: [%s] != [%s]", matchedTe.Time, expectedLocationTimestamp)
    } else if len(matchedTe.Items) != 1 {
        t.Fatalf("Expected exactly one location item to be matched: %v\n", matchedTe.Items)
    }

    gr := matchedTe.Items[0].(geoindex.GeographicRecord)

    expectedLatitude := float64(2.3)
    if gr.Latitude != expectedLatitude {
        t.Fatalf("Matched latitude not correct: [%.10f] != [%.10f]", gr.Latitude, expectedLatitude)
    }

    expectedLongitude := float64(20.3)
    if gr.Longitude != expectedLongitude {
        t.Fatalf("Matched longitude not correct: [%.10f] != [%.10f]", gr.Longitude, expectedLongitude)
    }
}

func TestFindGroups_FindLocationByTime_JustAfterLocationRecord(t *testing.T) {
    timeBase := epochUtc
    locationIndex := getTestLocationIndex(timeBase)

    fg := NewFindGroups(locationIndex, nil, nil)
    
    imageTimestamp := timeBase.Add(time.Hour * 1 + time.Minute * 11)
    
    imageTe := timeindex.TimeEntry{
        Time: imageTimestamp,
        Items: nil,
    }

    matchedTe, err := fg.findLocationByTime(imageTe)
    log.PanicIf(err)

    expectedLocationTimestamp := timeBase.Add(time.Hour * 1 + time.Minute * 10)

    if matchedTe.Time != expectedLocationTimestamp {
        t.Fatalf("The matched location timestamp is not correct: [%s] != [%s]", matchedTe.Time, expectedLocationTimestamp)
    } else if len(matchedTe.Items) != 1 {
        t.Fatalf("Expected exactly one location item to be matched: %v\n", matchedTe.Items)
    }

    gr := matchedTe.Items[0].(geoindex.GeographicRecord)

    expectedLatitude := float64(2.3)
    if gr.Latitude != expectedLatitude {
        t.Fatalf("Matched latitude not correct: [%.10f] != [%.10f]", gr.Latitude, expectedLatitude)
    }

    expectedLongitude := float64(20.3)
    if gr.Longitude != expectedLongitude {
        t.Fatalf("Matched longitude not correct: [%.10f] != [%.10f]", gr.Longitude, expectedLongitude)
    }
}

func TestFindGroups_FindLocationByTime_RoundUpToLocationRecord(t *testing.T) {
    timeBase := epochUtc
    locationIndex := getTestLocationIndex(timeBase)

    fg := NewFindGroups(locationIndex, nil, nil)
    
    imageTimestamp := timeBase.Add(time.Hour * 3 + time.Minute * 16)
    
    imageTe := timeindex.TimeEntry{
        Time: imageTimestamp,
        Items: nil,
    }

    matchedTe, err := fg.findLocationByTime(imageTe)
    log.PanicIf(err)

    expectedLocationTimestamp := timeBase.Add(time.Hour * 3 + time.Minute * 20)

    if matchedTe.Time != expectedLocationTimestamp {
        t.Fatalf("The matched location timestamp is not correct: [%s] != [%s]", matchedTe.Time, expectedLocationTimestamp)
    } else if len(matchedTe.Items) != 1 {
        t.Fatalf("Expected exactly one location item to be matched: %v\n", matchedTe.Items)
    }

    gr := matchedTe.Items[0].(geoindex.GeographicRecord)

    expectedLatitude := float64(4.3)
    if gr.Latitude != expectedLatitude {
        t.Fatalf("Matched latitude not correct: [%.10f] != [%.10f]", gr.Latitude, expectedLatitude)
    }

    expectedLongitude := float64(40.3)
    if gr.Longitude != expectedLongitude {
        t.Fatalf("Matched longitude not correct: [%.10f] != [%.10f]", gr.Longitude, expectedLongitude)
    }
}

func TestFindGroups_FindLocationByTime_RoundDownToLocationRecord(t *testing.T) {
    timeBase := epochUtc
    locationIndex := getTestLocationIndex(timeBase)

    fg := NewFindGroups(locationIndex, nil, nil)
    
    imageTimestamp := timeBase.Add(time.Hour * 3 + time.Minute * 14)
    
    imageTe := timeindex.TimeEntry{
        Time: imageTimestamp,
        Items: nil,
    }

    matchedTe, err := fg.findLocationByTime(imageTe)
    log.PanicIf(err)

    expectedLocationTimestamp := timeBase.Add(time.Hour * 3 + time.Minute * 10)

    if matchedTe.Time != expectedLocationTimestamp {
        t.Fatalf("The matched location timestamp is not correct: [%s] != [%s]", matchedTe.Time, expectedLocationTimestamp)
    } else if len(matchedTe.Items) != 1 {
        t.Fatalf("Expected exactly one location item to be matched: %v\n", matchedTe.Items)
    }

    gr := matchedTe.Items[0].(geoindex.GeographicRecord)

    expectedLatitude := float64(4.2)
    if gr.Latitude != expectedLatitude {
        t.Fatalf("Matched latitude not correct: [%.10f] != [%.10f]", gr.Latitude, expectedLatitude)
    }

    expectedLongitude := float64(40.2)
    if gr.Longitude != expectedLongitude {
        t.Fatalf("Matched longitude not correct: [%.10f] != [%.10f]", gr.Longitude, expectedLongitude)
    }
}

func TestFindGroups_FindLocationByTime_NoMatch(t *testing.T) {
    timeBase := epochUtc
    locationIndex := getTestLocationIndex(timeBase)

    fg := NewFindGroups(locationIndex, nil, nil)
    
    imageTimestamp := timeBase.Add(oneDay * 4 + time.Hour * 0 + time.Minute * 0)
    
    imageTe := timeindex.TimeEntry{
        Time: imageTimestamp,
        Items: nil,
    }

    _, err := fg.findLocationByTime(imageTe)
    if err != ErrNoNearLocationRecord {
        t.Fatalf("Didn't get error as expected for no matched location.")
    }
}

func getTestImageIndex(timeBase time.Time) (imageIndex *geoindex.Index) {
    imageIndex = geoindex.NewIndex()

    // Chicago
    chicagoCoordinates := []float64 { 41.85003, -87.65005 }

    // Detroit
    detroitCoordinates := []float64 { 42.33143, -83.04575 }

    // NYC
    nycCoordinates := []float64 { 40.71427, -74.00597 }

    // Sydney
    sydneyCoordinates := []float64 { -33.86785, 151.20732 }

    // Johannesburg
    joCoordinates := []float64 { -26.20227, 28.04363 }

    // Dresden
    dresdenCoordinates := []float64 { 51.05089, 13.73832 }

    // Note that we also mess-up the order in order to test that it's internally 
    // sorted.

    timeSeries := map[string]struct {
        timestamp time.Time
        latitude float64
        longitude float64
    } {
        "file01.jpg": { timeBase.Add(time.Hour * 0 + time.Minute * 1), chicagoCoordinates[0], chicagoCoordinates[1] },
        "file00.jpg": { timeBase.Add(time.Hour * 0 + time.Minute * 0), chicagoCoordinates[0], chicagoCoordinates[1] },
        "file04.jpg": { timeBase.Add(time.Hour * 0 + time.Minute * 4), chicagoCoordinates[0], chicagoCoordinates[1] },
        "file03.jpg": { timeBase.Add(time.Hour * 0 + time.Minute * 3), chicagoCoordinates[0], chicagoCoordinates[1] },
        "file02.jpg": { timeBase.Add(time.Hour * 0 + time.Minute * 2), chicagoCoordinates[0], chicagoCoordinates[1] },

        "file11.jpg": { timeBase.Add(time.Hour * 1 + time.Minute * 5), detroitCoordinates[0], detroitCoordinates[1] },
        "file10.jpg": { timeBase.Add(time.Hour * 1 + time.Minute * 0), detroitCoordinates[0], detroitCoordinates[1] },
        "file14.jpg": { timeBase.Add(time.Hour * 1 + time.Minute * 20), detroitCoordinates[0], detroitCoordinates[1] },
        "file13.jpg": { timeBase.Add(time.Hour * 1 + time.Minute * 15), detroitCoordinates[0], detroitCoordinates[1] },
        "file12.jpg": { timeBase.Add(time.Hour * 1 + time.Minute * 10), detroitCoordinates[0], detroitCoordinates[1] },

        "file21.jpg": { timeBase.Add(time.Hour * 2 + time.Minute * 1), nycCoordinates[0], nycCoordinates[1] },
        "file20.jpg": { timeBase.Add(time.Hour * 2 + time.Minute * 0), nycCoordinates[0], nycCoordinates[1] },
        "file24.jpg": { timeBase.Add(time.Hour * 2 + time.Minute * 4), nycCoordinates[0], nycCoordinates[1] },
        "file23.jpg": { timeBase.Add(time.Hour * 2 + time.Minute * 3), nycCoordinates[0], nycCoordinates[1] },
        "file22.jpg": { timeBase.Add(time.Hour * 2 + time.Minute * 2), nycCoordinates[0], nycCoordinates[1] },

        "file31.jpg": { timeBase.Add(time.Hour * 3 + time.Minute * 10), sydneyCoordinates[0], sydneyCoordinates[1] },
        "file30.jpg": { timeBase.Add(time.Hour * 3 + time.Minute * 0), sydneyCoordinates[0], sydneyCoordinates[1] },
        "file34.jpg": { timeBase.Add(time.Hour * 3 + time.Minute * 40), sydneyCoordinates[0], sydneyCoordinates[1] },
        "file33.jpg": { timeBase.Add(time.Hour * 3 + time.Minute * 30), sydneyCoordinates[0], sydneyCoordinates[1] },
        "file32.jpg": { timeBase.Add(time.Hour * 3 + time.Minute * 20), sydneyCoordinates[0], sydneyCoordinates[1] },

        "file41.jpg": { timeBase.Add(oneDay * 2 + time.Hour * 0 + time.Minute * 10), joCoordinates[0], joCoordinates[1] },
        "file40.jpg": { timeBase.Add(oneDay * 2 + time.Hour * 0 + time.Minute * 0), joCoordinates[0], joCoordinates[1] },
        "file44.jpg": { timeBase.Add(oneDay * 2 + time.Hour * 0 + time.Minute * 40), joCoordinates[0], joCoordinates[1] },
        "file43.jpg": { timeBase.Add(oneDay * 2 + time.Hour * 0 + time.Minute * 30), joCoordinates[0], joCoordinates[1] },
        "file42.jpg": { timeBase.Add(oneDay * 2 + time.Hour * 0 + time.Minute * 20), joCoordinates[0], joCoordinates[1] },

        "file51.jpg": { timeBase.Add(oneDay * 6 + time.Hour * 0 + time.Minute * 10), dresdenCoordinates[0], dresdenCoordinates[1] },
        "file50.jpg": { timeBase.Add(oneDay * 6 + time.Hour * 0 + time.Minute * 0), dresdenCoordinates[0], dresdenCoordinates[1] },
        "file54.jpg": { timeBase.Add(oneDay * 6 + time.Hour * 0 + time.Minute * 40), dresdenCoordinates[0], dresdenCoordinates[1] },
        "file53.jpg": { timeBase.Add(oneDay * 6 + time.Hour * 0 + time.Minute * 30), dresdenCoordinates[0], dresdenCoordinates[1] },
        "file52.jpg": { timeBase.Add(oneDay * 6 + time.Hour * 0 + time.Minute * 20), dresdenCoordinates[0], dresdenCoordinates[1] },
    }

    for filepath, x := range timeSeries {
        cameraModel := "some model"

        im := geoindex.ImageMetadata{
            CameraModel: cameraModel,
        }

        imageIndex.Add(geoindex.SourceImageJpeg, filepath, x.timestamp, true, x.latitude, x.longitude, 0, im)
    }

    return imageIndex
}

// TODO(dustin): !! Still need to test camera-model grouping.

func getCityIndex(cityDataFilepath string) *geoattractorindex.CityIndex {
    defer func() {
        if state := recover(); state != nil {
            err := log.Wrap(state.(error))
            log.PrintError(err)
            panic(err)
        }
    }()

    // Load countries.

    countryDataFilepath := path.Join(testAssetsPath, "countryInfo.txt")

    f, err := os.Open(countryDataFilepath)
    log.PanicIf(err)

    defer f.Close()

    countries, err := geoattractorparse.BuildGeonamesCountryMapping(f)
    log.PanicIf(err)

    // Load cities.

    gp := geoattractorparse.NewGeonamesParser(countries)

    g, err := os.Open(cityDataFilepath)
    log.PanicIf(err)

    defer g.Close()

    ci := geoattractorindex.NewCityIndex()

    err = ci.Load(gp, g)
    log.PanicIf(err)

    return ci
}

func checkGroup(t *testing.T, fg *FindGroups, finishedGroupKey GroupKey, finishedGroup []geoindex.GeographicRecord, expectedTimeKey time.Time, expectedCountry, expectedCity string, expectedFilenames []string) {
    if finishedGroupKey.TimeKey != expectedTimeKey {
        t.Fatalf("Time-key not correct: [%s] != [%s]\n", finishedGroupKey.TimeKey, expectedTimeKey)
    }

    cityLookup := fg.NearestCityIndex()
    cityRecord := cityLookup[finishedGroupKey.NearestCityKey]
    if cityRecord.Country != expectedCountry || cityRecord.City != expectedCity {
        t.Fatalf("Matched city not correct: %s", cityRecord)
    }

    if finishedGroupKey.ExifCameraModel != "some model" {
        t.Fatalf("Camera model not correct: [%s]", finishedGroupKey.ExifCameraModel)
    }

    if len(finishedGroup) != len(expectedFilenames) {
        t.Fatalf("Group is not the right size: (%d) != (%d)", len(finishedGroup), len(expectedFilenames))
    }

    for i, gr := range finishedGroup {
        if gr.Filepath != expectedFilenames[i] {
            for j, actualGr := range finishedGroup {
                fmt.Printf("(%d): [%s]\n", j, actualGr.Filepath)
            }

            t.Fatalf("File-path (%d) in group is not correct: [%s] != [%s]", i, gr.Filepath, expectedFilenames[i])
        }
    }
}

func TestFindGroups_FindNext_ImagesWithLocations(t *testing.T) {
    defer func() {
        if state := recover(); state != nil {
            err := log.Wrap(state.(error))
            log.PrintError(err)

            panic(err)
        }
    }()
    
    // locationIndex is just a non-empty index. We won't use it, but it needs to 
    // be present with at least one entry.
    locationIndex := geoindex.NewIndex()

    locationIndex.Add("some source", "file1", epochUtc, true, 1.1, 10.1, 0, nil)

    timeBase := epochUtc
    imageIndex := getTestImageIndex(timeBase)

    cityDataFilepath := path.Join(testAssetsPath, "allCountries.txt.multiple_major_cities_handpicked")
    ci := getCityIndex(cityDataFilepath)

    fg := NewFindGroups(locationIndex, imageIndex, ci)

    finishedGroupKey, finishedGroup, err := fg.FindNext()
    log.PanicIf(err)

    alignedTimeKey := timeBase.Add(time.Hour * 0 + time.Minute * 0)

    checkGroup(
        t, fg, 
        finishedGroupKey, 
        finishedGroup, 
        alignedTimeKey, 
        "United States", "Chicago", 
        []string { "file00.jpg", "file01.jpg", "file02.jpg", "file03.jpg", "file04.jpg" })

    finishedGroupKey, finishedGroup, err = fg.FindNext()
    log.PanicIf(err)

    // Same time-key but different city.
    checkGroup(
        t, fg, 
        finishedGroupKey, 
        finishedGroup, 
        alignedTimeKey, 
        "United States", "Detroit", 
        []string { "file10.jpg", "file11.jpg", "file12.jpg", "file13.jpg", "file14.jpg" })

    finishedGroupKey, finishedGroup, err = fg.FindNext()
    log.PanicIf(err)

    // Same time-key but different city.
    checkGroup(
        t, fg, 
        finishedGroupKey, 
        finishedGroup, 
        alignedTimeKey, 
        "United States", "New York City", 
        []string { "file20.jpg", "file21.jpg", "file22.jpg", "file23.jpg", "file24.jpg" })

    finishedGroupKey, finishedGroup, err = fg.FindNext()
    log.PanicIf(err)

    // Same time-key but different city.
    checkGroup(
        t, fg, 
        finishedGroupKey, 
        finishedGroup, 
        alignedTimeKey, 
        "Australia", "Sydney", 
        []string { "file30.jpg", "file31.jpg", "file32.jpg", "file33.jpg", "file34.jpg" })

    finishedGroupKey, finishedGroup, err = fg.FindNext()
    log.PanicIf(err)

    alignedTimeKey = timeBase.Add(oneDay * 2 + time.Hour * 0 + time.Minute * 0)

    checkGroup(
        t, fg, 
        finishedGroupKey, 
        finishedGroup, 
        alignedTimeKey, 
        "South Africa", "Johannesburg", 
        []string { "file40.jpg", "file41.jpg", "file42.jpg", "file43.jpg", "file44.jpg" })

    finishedGroupKey, finishedGroup, err = fg.FindNext()
    log.PanicIf(err)

    alignedTimeKey = timeBase.Add(oneDay * 6 + time.Hour * 0 + time.Minute * 0)

    checkGroup(
        t, fg, 
        finishedGroupKey, 
        finishedGroup, 
        alignedTimeKey, 
        "Germany", "Dresden", 
        []string { "file50.jpg", "file51.jpg", "file52.jpg", "file53.jpg", "file54.jpg" })

    _, _, err = fg.FindNext()
    if err != ErrNoMoreGroups {
        t.Fatalf("Expected no-more-groups error.")
    }
}
