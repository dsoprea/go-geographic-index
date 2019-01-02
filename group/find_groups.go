package geogroup

import (
    "errors"
    "fmt"
    "time"

    "github.com/dsoprea/go-logging"
    "github.com/dsoprea/go-time-index"
    "github.com/dsoprea/go-geographic-index"
    "github.com/dsoprea/go-geographic-attractor/index"
    "github.com/dsoprea/go-geographic-attractor"
)

var (
    ErrNoMoreGroups = errors.New("no more groups")
    ErrNoNearLocationRecord = errors.New("no location record was near-enough")
)

const (
    // DefaultRoundingWindowDuration is the largest time duration we're allowed
    // to search for matching location records within for a given image.
    DefaultRoundingWindowDuration = time.Minute * 10

    // DefaultCoalescenceWindowDuration is the distance that we'll use to
    // determine if the current image might belong to the same group as the last
    // image if all of the other factors match.
    DefaultCoalescenceWindowDuration = time.Hour * 24
)

const (
    SkipReasonNoNearLocationRecord = "no matching/near location record"
    SkipReasonNoNearCity           = "no near city"
)

var (
    findGroupsLogger = log.NewLogger("geogroup.find_groups")
)

type UnassignedRecord struct {
    Geographic geoindex.GeographicRecord
    Reason     string
}

type groupKey struct {
    TimeKey         time.Time
    NearestCityKey  string
    ExifCameraModel string
}

func (gk groupKey) String() string {
    return fmt.Sprintf("GroupKey<TIME-KEY=[%s] NEAREST-CITY=[%s] CAMERA-MODEL=[%s]>", gk.TimeKey, gk.NearestCityKey, gk.ExifCameraModel)
}

type FindGroups struct {
    locationIndex        *geoindex.Index
    imageIndex           *geoindex.Index
    unassignedRecords              []UnassignedRecord
    currentImagePosition int
    cityIndex            *geoattractorindex.CityIndex
    nearestCityIndex     map[string]geoattractor.CityRecord
    currentGroupKey      groupKey
    currentGroup         []geoindex.GeographicRecord

    roundingWindowDuration    time.Duration
    coalescenceWindowDuration time.Duration
}

func NewFindGroups(locationIndex *geoindex.Index, imageIndex *geoindex.Index, ci *geoattractorindex.CityIndex) *FindGroups {
    return &FindGroups{
        locationIndex:             locationIndex,
        imageIndex:                imageIndex,
        unassignedRecords:                   make([]UnassignedRecord, 0),
        cityIndex:                 ci,
        nearestCityIndex:          make(map[string]geoattractor.CityRecord),
        currentGroup:              make([]geoindex.GeographicRecord, 0),
        roundingWindowDuration:    DefaultRoundingWindowDuration,
        coalescenceWindowDuration: DefaultCoalescenceWindowDuration,
    }
}

func (fg *FindGroups) SetRoundingWindowDuration(roundingWindowDuration time.Duration) {
    fg.roundingWindowDuration = roundingWindowDuration
}

func (fg *FindGroups) SetCoalescenceWindowDuration(coalescenceWindowDuration time.Duration) {
    fg.coalescenceWindowDuration = coalescenceWindowDuration
}

// NearestCityIndex returns all of the cities that we've grouped the images by
// in a map keyed the same as in the grouping.
func (fg *FindGroups) NearestCityIndex() map[string]geoattractor.CityRecord {
    return fg.nearestCityIndex
}

func (fg *FindGroups) UnassignedRecords() []UnassignedRecord {
    return fg.unassignedRecords
}

func (fg *FindGroups) addUnassigned(gr geoindex.GeographicRecord, reason string) {
    ur := UnassignedRecord{
        Geographic: gr,
        Reason:     reason,
    }

    fg.unassignedRecords = append(fg.unassignedRecords, ur)

    findGroupsLogger.Warningf(nil, "Skipping %s: %s", gr, reason)
}

// findLocationByTime returns the nearest location record to the timestamp in 
// the given image record.
func (fg *FindGroups) findLocationByTime(imageTe timeindex.TimeEntry) (matchedTe timeindex.TimeEntry, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    locationIndexTs := fg.locationIndex.Series()

    // nearestLocationPosition is either the position where the exact
    // time of the image was found in the location index or the
    // position that it would be inserted (even though we're not
    // interested in insertions).
    //
    // Both the location and image indices are ordered, obviously;
    // technically we could potentially read along both and avoid a
    // bunch of bunch searches. However, the location index will be
    // frequented by large gaps that have no corresponding images and
    // we're just going to end-up seeking more that way.
    nearestLocationPosition := timeindex.SearchTimes(locationIndexTs, imageTe.Time)

    var previousLocationTe timeindex.TimeEntry
    var nextLocationTe timeindex.TimeEntry

    if nearestLocationPosition >= len(locationIndexTs) {
        // We were given a position past the end of the list.

        previousLocationTe = locationIndexTs[len(locationIndexTs)-1]
    } else {
        // We were given a position within the list.

        nearestLocationTe := locationIndexTs[nearestLocationPosition]
        if nearestLocationTe.Time == imageTe.Time {
            // We found a location record that exactly matched our
            // image record (time-wise).

            return nearestLocationTe, nil
        } else {
            // This is an optimistic insertion-position recommendation
            // (`nearestLocationPosition` is a existing record that is
            // larger than our query).

            nextLocationTe = nearestLocationTe
        }

        // If there's at least one more entry to the left,
        // calculate the distance to it.
        if nearestLocationPosition > 0 {
            previousLocationTe = locationIndexTs[nearestLocationPosition-1]
        }
    }

    var durationSincePrevious time.Duration
    if previousLocationTe.IsZero() == false {
        durationSincePrevious = imageTe.Time.Sub(previousLocationTe.Time)
    }

    var durationUntilNext time.Duration
    if nextLocationTe.IsZero() == false {
        durationUntilNext = nextLocationTe.Time.Sub(imageTe.Time)
    }

    if durationSincePrevious != 0 {
        if durationSincePrevious <= fg.roundingWindowDuration && (durationUntilNext == 0 || durationUntilNext > fg.roundingWindowDuration) {
            // Only the preceding time duration is acceptable.
            matchedTe = previousLocationTe
        } else if durationSincePrevious <= fg.roundingWindowDuration && durationUntilNext != 0 && durationUntilNext <= fg.roundingWindowDuration {
            // They're both fine. Take the nearest.

            if durationSincePrevious < durationUntilNext {
                matchedTe = previousLocationTe
            } else {
                matchedTe = nextLocationTe
            }
        }
    }

    // Effectively, the "else" for the above.
    if durationUntilNext != 0 && matchedTe.IsZero() == true && durationUntilNext < fg.roundingWindowDuration {
        matchedTe = nextLocationTe
    }

    if matchedTe.Time.IsZero() == true {
        return timeindex.TimeEntry{}, ErrNoNearLocationRecord
    }

    return matchedTe, nil
}

func (fg *FindGroups) flushCurrentGroup(nextGroupKey groupKey) (finishedGroupKey groupKey, finishedGroup []geoindex.GeographicRecord, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    finishedGroup = fg.currentGroup
    fg.currentGroup = make([]geoindex.GeographicRecord, 0)

    finishedGroupKey = fg.currentGroupKey
    fg.currentGroupKey = nextGroupKey

    return finishedGroupKey, finishedGroup, nil
}

// FindNext returns the next set of grouped-images along with the actual
// grouping factors.
func (fg *FindGroups) FindNext() (finishedGroupKey groupKey, finishedGroup []geoindex.GeographicRecord, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if len(fg.locationIndex.Series()) == 0 {
        log.Panicf("no locations in index")
    }

    imageIndexTs := fg.imageIndex.Series()

    if fg.currentImagePosition >= len(imageIndexTs) {
        return groupKey{}, nil, ErrNoMoreGroups
    }

    for ; fg.currentImagePosition < len(imageIndexTs); fg.currentImagePosition++ {
        imageTe := imageIndexTs[fg.currentImagePosition]
        for _, item := range imageTe.Items {
            imageGr := item.(geoindex.GeographicRecord)

            latitude := imageGr.Latitude
            longitude := imageGr.Longitude

            if imageGr.HasGeographic == false {
                matchedTe, err := fg.findLocationByTime(imageTe)
                if err != nil {
                    if log.Is(err, ErrNoNearLocationRecord) == true {
                        fg.addUnassigned(imageGr, SkipReasonNoNearLocationRecord)
                        continue
                    }

                    log.Panic(err)
                }

                locationItem := matchedTe.Items[0]
                locationGr := locationItem.(geoindex.GeographicRecord)

                // The location index should exclusively be loaded with
                // geographic data. This should never happen.
                if locationGr.HasGeographic == false {
                    log.Panicf("location record indicates no geographic data; this should never happen")
                }

                latitude = locationGr.Latitude
                longitude = locationGr.Longitude
            }

            // If we got here, we either have or have found a location for the
            // given image.

            // Now, we'll construct the group that this image should be a part
            // of. Later, we'll compare the groups of each image to the groups
            // of adjacent images in order to determine which should be binned
            // together.

            // First, find a city to associate this location with.

            sourceName, _, cr, err := fg.cityIndex.Nearest(latitude, longitude)
            if err != nil {
                if log.Is(err, geoattractorindex.ErrNoNearestCity) == true {
                    fg.addUnassigned(imageGr, SkipReasonNoNearCity)
                    continue
                }

                log.Panic(err)
            }

            nearestCityKey := fmt.Sprintf("%s,%s", sourceName, cr.Id)
            fg.nearestCityIndex[nearestCityKey] = cr

            // Determine what timestamp to associate this image to.

            imageUnixTime := imageTe.Time.Unix()
            timeKey := time.Unix(imageUnixTime - imageUnixTime%10, 0)

            // If the current image's time is within X duration of the last
            // group's time-key, use the same time-key. If the last group's
            // other factors also match, this group will end up being the same.
            if fg.currentGroupKey.TimeKey.IsZero() == false && timeKey.Sub(fg.currentGroupKey.TimeKey) < fg.coalescenceWindowDuration {
                timeKey = fg.currentGroupKey.TimeKey
            }

            // Build the group key.

            gk := groupKey{
                TimeKey:        timeKey,
                NearestCityKey: nearestCityKey,
            }

            if imageGr.SourceName == geoindex.SourceImageJpeg {
                metadata := imageGr.Metadata.(geoindex.ImageMetadata)
                gk.ExifCameraModel = metadata.CameraModel
            }

            // TODO(dustin): !! Let's keep an index of the last group seen for *each camera*.
            if gk != fg.currentGroupKey {
                finishedGroupKey, finishedGroup, err = fg.flushCurrentGroup(gk)
                log.PanicIf(err)

                fg.currentGroup = append(fg.currentGroup, imageGr)

                return finishedGroupKey, finishedGroup, nil
            } else {
                fg.currentGroup = append(fg.currentGroup, imageGr)
            }
        }
    }

    if len(fg.currentGroup) > 0 {
        finishedGroupKey, finishedGroup, err = fg.flushCurrentGroup(groupKey{})
        log.PanicIf(err)

        return finishedGroupKey, finishedGroup, nil
    }

    return groupKey{}, nil, ErrNoMoreGroups
}
