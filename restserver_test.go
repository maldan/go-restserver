package restserver

import (
	"testing"
)

func TestGetFile(t *testing.T) {
	f := getFile("package.json")
	if f == nil {
		t.Fatal("Get file error")
	}
}
