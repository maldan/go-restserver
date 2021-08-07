package restserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

func ErrorMessage(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")

	if err := recover(); err != nil {
		switch e := err.(type) {
		case *RestServerError:
			if e.StatusCode == 0 {
				rw.WriteHeader(500)
			} else {
				rw.WriteHeader(e.StatusCode)
			}

			responseData := RestServerResponse{Status: false, Error: RestServerError{StatusCode: e.StatusCode, Field: e.Field, Type: e.Type, Description: e.Description}}
			finalData, _ := json.Marshal(responseData)
			fmt.Fprintf(rw, "%+v", string(finalData))
		case error:
			rw.WriteHeader(500)

			responseData := RestServerResponse{Status: false, Error: RestServerError{StatusCode: 500, Type: ErrorType.Unknown, Description: e.Error()}}
			finalData, _ := json.Marshal(responseData)
			fmt.Fprintf(rw, "%+v", string(finalData))
		default:
			rw.WriteHeader(500)

			responseData := RestServerResponse{Status: false, Error: RestServerError{StatusCode: 500, Type: ErrorType.Unknown, Description: fmt.Sprintf("%v", e)}}
			finalData, _ := json.Marshal(responseData)
			fmt.Fprintf(rw, "%+v", string(finalData))
		}
	}
}

func ErrorWsMessage(conn *websocket.Conn, messageType int, msgId string) {
	if err := recover(); err != nil {
		switch e := err.(type) {
		default:
			/*rw.WriteHeader(500)

			responseData := RestServerResponse{Status: false, Error: RestServerError{StatusCode: 500, Type: ErrorType.Unknown, Description: fmt.Sprintf("%v", e)}}
			finalData, _ := json.Marshal(responseData)
			fmt.Fprintf(rw, "%+v", string(finalData))*/

			realOut, _ := json.Marshal(WsResponse{
				Id:       msgId,
				Status:   false,
				Response: fmt.Sprintf("%v", e),
			})
			conn.WriteMessage(messageType, realOut)
		}
	}
}

func Fatal(code int, kind string, field string, description string) {
	panic(&RestServerError{StatusCode: code, Type: kind, Field: field, Description: description})
}
