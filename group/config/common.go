package geogroupconfig

import (
    "os"
    "path"
)

var (
    appPath string
)

func init() {
    goPath := os.Getenv("GOPATH")
    appPath = path.Join(goPath, "src", "github.com", "dsoprea", "go-geographic-index", "group")
}
