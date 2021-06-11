package restserver

import (
	"encoding/json"
	"fmt"
	"net/http"
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

			responseData := RestResponse{Status: false, Description: e.Description}
			finalData, _ := json.Marshal(responseData)
			fmt.Fprintf(rw, "%+v", string(finalData))
		case error:
			rw.WriteHeader(500)

			responseData := RestResponse{Status: false, Description: e.Error()}
			finalData, _ := json.Marshal(responseData)
			fmt.Fprintf(rw, "%+v", string(finalData))
		default:
			rw.WriteHeader(500)

			responseData := RestResponse{Status: false, Description: e}
			finalData, _ := json.Marshal(responseData)
			fmt.Fprintf(rw, "%+v", string(finalData))
		}
	}
}

func Error(code int, description string) {
	panic(&RestServerError{StatusCode: code, Description: description})
}
