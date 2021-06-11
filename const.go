package restserver

import "fmt"

type RestServerContext struct {
	ContentType string
	StatusCode  int
}

type RestResponse struct {
	Status      bool        `json:"status"`
	Description interface{} `json:"description"`
	Response    interface{} `json:"response"`
}

type RestServerError struct {
	StatusCode  int
	Description string
}

func (e *RestServerError) Error() string {
	return fmt.Sprintf("parse %v: internal error", e.Description)
}
