package restserver

import (
	"net/http"
	"os"
	"path"
)

func GetMimeByExt(ext string) string {
	contentType := "application/octet-stream"
	if ext == ".md" || ext == ".go" || ext == ".txt" {
		contentType = "text/plain; charset=utf-8"
	}
	if ext == ".html" {
		contentType = "text/html; charset=utf-8"
	}
	if ext == ".css" {
		contentType = "text/css; charset=utf-8"
	}
	if ext == ".js" {
		contentType = "text/javascript; charset=utf-8"
	}
	if ext == ".json" {
		contentType = "application/json; charset=utf-8"
	}
	if ext == ".svg" {
		contentType = "image/svg+xml"
	}
	if ext == ".mp4" {
		contentType = "video/mp4"
	}
	return contentType
}

func GetMimeByFile(file *os.File) (string, error) {
	// Get file header
	file.Seek(0, 0)
	buffer := make([]byte, 1024)
	_, err := file.Read(buffer)
	if err != nil {
		return "", err
	}

	// Detect content type
	contentType := http.DetectContentType(buffer)

	// Binary
	if contentType == "application/octet-stream" {
		if string(buffer[4:12]) == "ftypisom" {
			contentType = "video/mp4"
		}
	}

	/*ext := path.Ext(p)
	if contentType == "application/octet-stream" || contentType == "text/plain; charset=utf-8" {
		contentType = GetMimeByExt(ext)
	}*/

	return contentType, nil
}

// Get mime type
func GetMime(p string) (string, error) {
	// Open file
	file, err := os.Open(p)
	if err != nil {
		return "", err
	}

	// Get file header
	file.Seek(0, 0)
	buffer := make([]byte, 1024)
	_, err = file.Read(buffer)
	if err != nil {
		return "", err
	}

	// Detect content type
	contentType := http.DetectContentType(buffer)
	ext := path.Ext(p)
	if contentType == "application/octet-stream" || contentType == "text/plain; charset=utf-8" {
		contentType = GetMimeByExt(ext)
	}

	return contentType, nil
}
