package geoindex

import (
	"os"
	"path"
)

var (
	appPath        string
	testAssetsPath string
)

func init() {
	goPath := os.Getenv("GOPATH")
	appPath = path.Join(goPath, "src", "github.com", "dsoprea", "go-geographic-index")
	testAssetsPath = path.Join(appPath, "test", "asset")
}
