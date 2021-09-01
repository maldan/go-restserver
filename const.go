package restserver

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"

	"github.com/gorilla/websocket"
)

var ErrorType = struct {
	EmptyField   string
	AccessDenied string
	Unknown      string
	NotFound     string
}{
	EmptyField:   "emptyField",
	AccessDenied: "accessDenied",
	Unknown:      "unknown",
	NotFound:     "notFound",
}

type RestServerContext struct {
	ContentType string
	StatusCode  int
}

type RestServerResponse struct {
	Status   bool            `json:"status"`
	Error    RestServerError `json:"error"`
	Response interface{}     `json:"response"`
}

type RestServerError struct {
	StatusCode  int    `json:"code"`
	Type        string `json:"type"`
	Field       string `json:"field"`
	Description string `json:"description"`
}

type WsClient struct {
	Id         string
	Connection *websocket.Conn
}

type WsMessage struct {
	Id     string          `json:"id"`
	Method string          `json:"method"`
	Args   json.RawMessage `json:"args"`
	Data   []byte          `json:"data"`
}

type WsResponse struct {
	Id       string      `json:"id"`
	Status   bool        `json:"status"`
	Response interface{} `json:"response"`
}

type WsEvent struct {
	Event    string      `json:"event"`
	Response interface{} `json:"response"`
}

type VirtualFs struct {
	Root string
	Fs   embed.FS
}

// Send JSON event to client
func (c WsClient) SendEventJSON(event string, data interface{}) {
	b, _ := json.Marshal(WsEvent{
		Event:    event,
		Response: data,
	})
	c.Connection.WriteMessage(websocket.TextMessage, b)
}

func (e *RestServerError) Error() string {
	return fmt.Sprintf("parse %v: internal error", e.Description)
}

func (fs *VirtualFs) Open(path string) (fs.File, error) {
	return fs.Fs.Open(strings.ReplaceAll(fs.Root+path, "//", "/"))
}
