package rest_server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
)

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

func CallMethod2(controller interface{}, method reflect.Method, params map[string]interface{}, context *RestServerContext) (result reflect.Value, err error) {
	function := reflect.ValueOf(method.Func.Interface())
	functionType := reflect.TypeOf(method.Func.Interface())
	firstArgument := functionType.In(1)
	args := reflect.New(firstArgument).Interface()
	s := reflect.ValueOf(args).Elem()

	if s.Kind() == reflect.Struct {
		// Fill context
		contextField := s.FieldByName("Context")
		if contextField.IsValid() {
			if contextField.CanSet() {
				contextField.Set(reflect.ValueOf(context))
			}
		}

		for k, v := range params {
			f := s.FieldByName(strings.Title(k))

			if f.IsValid() {
				if f.CanSet() {
					// Fill string types
					if f.Kind() == reflect.String && reflect.TypeOf(v).Kind() == reflect.String {
						f.SetString(v.(string))
					}

					// Fill int types
					if f.Kind() == reflect.Int64 {
						// From int
						if reflect.TypeOf(v).Kind() == reflect.Int64 {
							f.SetInt(v.(int64))
						}
						// From string
						if reflect.TypeOf(v).Kind() == reflect.String {
							i, _ := strconv.ParseInt(v.(string), 10, 64)
							f.SetInt(i)
						}
					}

					// Fill int types
					if f.Kind() == reflect.Bool {
						// From int
						if reflect.TypeOf(v).Kind() == reflect.Bool {
							f.SetBool(v.(bool))
						}
						// From string
						if reflect.TypeOf(v).Kind() == reflect.String {
							if v.(string) == "true" {
								f.SetBool(true)
							} else {
								f.SetBool(false)
							}
						}
					}

					// Fill int types
					if f.Kind() == reflect.Slice {
						f.Set(reflect.ValueOf(v))
					}
				}
			}
		}
	}

	// Call function
	in := make([]reflect.Value, 2)
	in[0] = reflect.ValueOf(controller)
	in[1] = reflect.ValueOf(s.Interface())
	result = function.Call(in)[0]

	return
}

func CallMethod(controller interface{}, methodName string, params map[string]interface{}, context *RestServerContext) (result reflect.Value, err error) {
	fooType := reflect.TypeOf(controller)
	for i := 0; i < fooType.NumMethod(); i++ {
		method := fooType.Method(i)
		if methodName == method.Name {
			result, err = CallMethod2(controller, method, params, context)
			return
		}
	}
	Error(500, "Method not found")
	return
}

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

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func FileHandler(path string) *os.File {
	dir, err := os.Getwd()
	if err != nil {
		dir = ""
	}

	if FileExists(dir + path) {
		f, err := os.Open(dir + path)
		if err != nil {
			return nil
		}
		// defer f.Close()
		return f
	}
	return nil
}

func Start(addr string, controller map[string]interface{}) {
	fmt.Printf("Starting server at " + addr + "\n")

	// Set handler
	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		defer ErrorMessage(rw, r)

		// Fuck cors
		rw.Header().Add("Access-Control-Allow-Origin", "*")
		rw.Header().Add("Access-Control-Allow-Methods", "*")
		rw.Header().Add("Access-Control-allow-Headers", "*")

		// Fuck options
		if r.Method == "OPTIONS" {
			rw.WriteHeader(200)
			fmt.Fprintf(rw, "")
			return
		}

		// Collect args
		args := map[string]interface{}{
			"accessToken": r.Header.Get("Authorization"),
		}
		for key, element := range r.URL.Query() {
			args[key] = element[0]
		}

		// Parse body
		if strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
			// Parse multipart body and collect args
			r.ParseMultipartForm(0)
			for key, element := range r.MultipartForm.Value {
				args[key] = element[0]
			}
			if len(r.MultipartForm.File) > 0 {
				args["files"] = r.MultipartForm.File
			}
		} else {
			// Parse json body and collect args
			jsonMap := make(map[string]interface{})
			json.NewDecoder(r.Body).Decode(&jsonMap)
			for key, element := range jsonMap {
				args[key] = element
			}
		}

		// Check file and return if found
		file := FileHandler(r.URL.Path)
		if file != nil {
			stat, _ := file.Stat()

			// Get file header
			file.Seek(0, 0)
			buffer := make([]byte, 512)
			_, err := file.Read(buffer)
			if err != nil {
				Error(500, "Can't read file")
			}

			// Detect content type
			contentType := http.DetectContentType(buffer)
			ext := path.Ext(r.URL.Path)
			if contentType == "application/octet-stream" {
				if ext == ".md" {
					contentType = "text/plain; charset=utf-8"
				}
			}

			// Set headers
			rw.Header().Add("Content-Type", contentType)
			rw.Header().Add("Content-Length", fmt.Sprintf("%d", stat.Size()))

			// Stream file
			file.Seek(0, 0)
			io.Copy(rw, file)
			return
		}

		// Get controller
		path := strings.Split(r.URL.Path, "/")
		controllerName := path[1]
		methodName := path[2]

		// Check controller
		if controller[controllerName] == nil {
			Error(404, "Controller not found")
		}

		// Call method
		context := new(RestServerContext)
		context.ContentType = "application/json"
		context.StatusCode = 200
		response, err := CallMethod(controller[controllerName], strings.Title(strings.ToLower(r.Method))+strings.Title(methodName), args, context)

		if err != nil {
			Error(500, err.Error())
		}

		// Response
		rw.Header().Add("Content-Type", context.ContentType)
		if context.ContentType == "application/json" {
			responseData := RestResponse{Status: true}
			responseData.Response = response.Interface()
			finalData, _ := json.Marshal(responseData)
			fmt.Fprintf(rw, "%+v", string(finalData))
		} else {
			fmt.Fprintf(rw, "%+v", response)
		}
	})

	// Start server
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
		return
	}
}
