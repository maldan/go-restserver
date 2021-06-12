package restserver

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"
)

type RestServerContext struct {
	ContentType string
	StatusCode  int
}

type RestResponse struct {
	Status      bool        `json:"status"`
	Description interface{} `json:"description"`
	Response    interface{} `json:"response"`
}

type RestServerError struct {
	StatusCode  int
	Description string
}

type VirtualFs struct {
	Root string
	Fs   embed.FS
}

func (e *RestServerError) Error() string {
	return fmt.Sprintf("parse %v: internal error", e.Description)
}

func (fs *VirtualFs) Open(path string) (fs.File, error) {
	return fs.Fs.Open(strings.ReplaceAll(fs.Root+path, "//", "/"))
}
