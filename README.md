[![Build Status](https://travis-ci.org/dsoprea/go-geographic-index.svg?branch=master)](https://travis-ci.org/dsoprea/go-geographic-index)
[![Coverage Status](https://coveralls.io/repos/github/dsoprea/go-geographic-index/badge.svg?branch=master)](https://coveralls.io/github/dsoprea/go-geographic-index?branch=master)
[![GoDoc](https://godoc.org/github.com/dsoprea/go-geographic-index?status.svg)](https://godoc.org/github.com/dsoprea/go-geographic-index)


# Overview

An in-memory time-series index that can be loaded from recursively processing directories of images and GPX files. In the case of image files, we will attempt to extract a coordinate from GPS information in the EXIF (if present).

**This is the underlying storage mechanism of the image-grouping tool in [group/](https://github.com/dsoprea/go-geographic-index/tree/master/group).**


# Examples

Records can be added to the index either directly or automatically from recursively processing a given path and extracting locations from GPX and JPEG files:

[GeographicCollector.ReadFromPath](https://godoc.org/github.com/dsoprea/go-geographic-index#example-GeographicCollector-ReadFromPath)

```go
index := NewIndex()
gc := NewGeographicCollector(index)

err := RegisterImageFileProcessors(gc)
log.PanicIf(err)

err = RegisterDataFileProcessors(gc)
log.PanicIf(err)

err = gc.ReadFromPath(testAssetsPath)
log.PanicIf(err)
```

[Index.Add](https://godoc.org/github.com/dsoprea/go-geographic-index#example-Index-Add)

```go
index := NewIndex()

epochUtc := (time.Time{}).UTC()
hasGeographic := true
latitude := float64(123.456)
longitude := float64(789.012)
var metadata interface{}

index.Add(SourceGeographicGpx, "data.gpx", epochUtc, hasGeographic, latitude, longitude, metadata)
```

The ordered index data can also be exported back to a GPX file:

[Index.ExportGpx](https://godoc.org/github.com/dsoprea/go-geographic-index#example-Index-ExportGpx)

```go
index := NewIndex()
gc := NewGeographicCollector(index)

err := RegisterImageFileProcessors(gc)
log.PanicIf(err)

err = RegisterDataFileProcessors(gc)
log.PanicIf(err)

err = gc.ReadFromPath(testAssetsPath)
log.PanicIf(err)

buffer := new(bytes.Buffer)

err = gc.index.ExportGpx(buffer)
log.PanicIf(err)
```