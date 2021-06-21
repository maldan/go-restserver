package restserver

import (
	"os"
	"unicode"
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

func lowerFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}
