package main

import (
	"fmt"
	"time"

	"github.com/maldan/go-restserver"
)

type TestApi int

type XArgs struct {
	A string `validation:"required" json:"sas"`
}

type FuckArgs struct {
	City    string
	Sos     map[string]string
	Created time.Time
}

type AAArgs struct {
	A int `json:"a"`
	B int `json:"b"`
}

func (u TestApi) WsSasageo(args AAArgs) int {
	return args.A + args.B
}

func (u TestApi) GetSasageo(args XArgs) string {
	return fmt.Sprintf("%#+v", args)
}

func (u TestApi) PostFuck(args FuckArgs) string {
	fmt.Printf("%#+v\n", args)
	return "X"
}

func main() {
	restserver.Start("127.0.0.1:9512", map[string]interface{}{
		"/": "/",
		"/api": map[string]interface{}{
			"test": new(TestApi),
		},
		"/ws": map[string]interface{}{
			"test": new(TestApi),
		},
	})
}
