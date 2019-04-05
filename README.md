[![Build Status](https://travis-ci.org/dsoprea/go-geographic-index.svg?branch=master)](https://travis-ci.org/dsoprea/go-geographic-index)
[![Coverage Status](https://coveralls.io/repos/github/dsoprea/go-geographic-index/badge.svg?branch=master)](https://coveralls.io/github/dsoprea/go-geographic-index?branch=master)
[![GoDoc](https://godoc.org/github.com/dsoprea/go-geographic-index?status.svg)](https://godoc.org/github.com/dsoprea/go-geographic-index)


# Overview

An in-memory time-series index that can be loaded manually, or automatically from image (using EXIF) and GPS data-log files. In the case of image files, we will attempt to extract a coordinate from GPS information in the EXIF (if present).

**This is the underlying storage mechanism of the [go-geographic-autogroup-images](https://github.com/dsoprea/go-geographic-autogroup-images) project.**


# Components

- [github.com/dsoprea/go-time-index](https://github.com/dsoprea/go-time-index)
- [github.com/dsoprea/go-jpeg-image-structure](https://github.com/dsoprea/go-jpeg-image-structure)
- [github.com/dsoprea/go-exif](https://github.com/dsoprea/go-exif)
- [github.com/dsoprea/go-gpx](https://github.com/dsoprea/go-gpx)


# Examples

Records can be added to the index either directly or automatically from recursively processing a given path and extracting locations from GPS data-log and image files (those supporting and having EXIF). Currently, only GPX files are supported for data-logs and JPEG files for images.

Excerpt from [GeographicCollector.ReadFromPath](https://godoc.org/github.com/dsoprea/go-geographic-index#example-GeographicCollector-ReadFromPath) example:

```go
index := NewTimeIndex()
gc := NewGeographicCollector(index)

err := RegisterImageFileProcessors(gc, 0, nil)
log.PanicIf(err)

err = RegisterDataFileProcessors(gc)
log.PanicIf(err)

err = gc.ReadFromPath(testAssetsPath)
log.PanicIf(err)
```

Excerpt from [Index.Add](https://godoc.org/github.com/dsoprea/go-geographic-index#example-Index-Add) example:

```go
index := NewTimeIndex()

epochUtc := (time.Time{}).UTC()
hasGeographic := true
latitude := float64(123.456)
longitude := float64(789.012)
var metadata interface{}

index.Add(SourceGeographicGpx, "data.gpx", epochUtc, hasGeographic, latitude, longitude, metadata)
```

The ordered index data can also be exported back to a GPX file:

Excerpt from [Index.ExportGpx](https://godoc.org/github.com/dsoprea/go-geographic-index#example-Index-ExportGpx) example:

```go
index := NewTimeIndex()
gc := NewGeographicCollector(index)

err := RegisterImageFileProcessors(gc, 0, nil)
log.PanicIf(err)

err = RegisterDataFileProcessors(gc)
log.PanicIf(err)

err = gc.ReadFromPath(testAssetsPath)
log.PanicIf(err)

buffer := new(bytes.Buffer)

err = gc.index.ExportGpx(buffer)
log.PanicIf(err)
```
