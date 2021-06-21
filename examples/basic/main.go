package main

import (
	"fmt"

	"github.com/maldan/go-restserver"
)

type TestApi int

type XArgs struct {
	A string `validation:"required" json:"sas"`
}

type FuckArgs struct {
	Type               string
	City               string
	Address            string
	Area               string
	SleepingPlaces     string
	RoomAmount         string
	PricePerDay        string
	WeekendPricePerDay string
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
	})
}
