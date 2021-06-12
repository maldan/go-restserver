package restserver

import (
	"os"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func getFile(path string) *os.File {
	if path[len(path)-1] == '/' {
		path += "index.html"
	}

	if fileExists(path) {
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		// defer f.Close()
		return f
	}
	return nil
}
