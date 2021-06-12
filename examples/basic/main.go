package main

import "github.com/maldan/go-restserver"

type TestApi int

func (u TestApi) GetSasageo() string {
	return "X"
}

func main() {
	restserver.Start("127.0.0.1:9512", map[string]interface{}{
		"/": "/",
		"/api": map[string]interface{}{
			"test": new(TestApi),
		},
	})
}
