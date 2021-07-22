package restserver

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"reflect"
	"strings"
	"time"
)

func FillFieldList(s *reflect.Value, ss reflect.Type, params map[string]interface{}) {
	amount := s.NumField()

	if params == nil {
		params = make(map[string]interface{})
	}

	for i := 0; i < amount; i++ {
		field := s.Field(i)
		fieldName := ss.Field(i).Name
		fieldTag := ss.Field(i).Tag
		jsonName := fieldTag.Get("json")

		// isRequired := fieldTag.Get("validation") == "required"

		// Can change field
		if field.IsValid() {
			if field.CanSet() {
				// Skip
				if jsonName == "-" {
					continue
				}

				// Get value
				var v interface{}
				if jsonName != "" {
					x, ok := params[jsonName]
					if x == nil {
						continue
					}
					if ok {
						v = x
					} else {
						continue
					}
				} else {
					x, ok := params[lowerFirst(fieldName)]
					if x == nil {
						continue
					}
					if ok {
						v = x
					} else {
						continue
					}
				}

				// Check
				/*if reflect.ValueOf(v).IsZero() && isRequired {
					Fatal(500, ErrorType.EmptyField, fieldName, fieldName+" is required")
				}*/

				// Get field type
				switch field.Kind() {
				case reflect.String:
					ApplyString(&field, v)
				case reflect.Uint64:
				case reflect.Uint32:
				case reflect.Uint16:
				case reflect.Uint8:
				case reflect.Uint:
				case reflect.Int64:
				case reflect.Int32:
				case reflect.Int16:
				case reflect.Int8:
				case reflect.Int:
					ApplyInt(&field, v)
				case reflect.Float32:
				case reflect.Float64:
					ApplyFloat(&field, v)
				case reflect.Bool:
					ApplyBool(&field, v)
				case reflect.Slice:
					ApplySlice(&field, v)
				case reflect.Struct:
					if field.Type().Name() == "Time" {
						ApplyTime(&field, v)
					} else {
						if reflect.TypeOf(v).Kind() == reflect.Map {
							FillFieldList(&field, reflect.TypeOf(field.Interface()), v.(map[string]interface{}))
						}
					}
				case reflect.Ptr:
					ApplyPtr(&field, v)
					continue
				default:
					continue
				}
			}
		}
	}
}

func CallMethod2(controller interface{}, method reflect.Method, params map[string]interface{}, context *RestServerContext) (result reflect.Value, err error) {
	function := reflect.ValueOf(method.Func.Interface())
	functionType := reflect.TypeOf(method.Func.Interface())

	// No args
	if functionType.NumIn() == 1 {
		in := make([]reflect.Value, 1)
		in[0] = reflect.ValueOf(controller)
		r := function.Call(in)
		if len(r) > 0 {
			result = r[0]
		} else {
			result = reflect.ValueOf("")
		}

		return
	}

	firstArgument := functionType.In(1)
	args := reflect.New(firstArgument).Interface()
	argsValue := reflect.ValueOf(args).Elem()
	argsType := reflect.TypeOf(args).Elem()

	// If first args is string
	if argsValue.Kind() == reflect.Struct {
		// Fill context
		contextField := argsValue.FieldByName("Context")
		if contextField.IsValid() {
			if contextField.CanSet() {
				contextField.Set(reflect.ValueOf(context))
			}
		}

		// Go over fields
		FillFieldList(&argsValue, argsType, params)
	}

	// Call function
	in := make([]reflect.Value, 2)
	in[0] = reflect.ValueOf(controller)
	in[1] = reflect.ValueOf(argsValue.Interface())
	r := function.Call(in)
	if len(r) > 0 {
		result = r[0]
	} else {
		result = reflect.ValueOf("")
	}

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
	Fatal(500, ErrorType.NotFound, "", "Method not found")
	return
}

func VirtualFileHandler(rw http.ResponseWriter, r *http.Request, fs VirtualFs) {
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

	// Check file and return if found
	finalPath := strings.ReplaceAll(r.URL.Path, "//", "/")
	if finalPath[len(finalPath)-1] == '/' {
		finalPath += "index.html"
	}

	file, err := fs.Open(finalPath)

	if err == nil {
		// stat, _ := file.Stat() //

		// Get file header
		buffer := make([]byte, 1024*1024*4)
		totalSize, err := file.Read(buffer)
		if err != nil {
			Fatal(500, ErrorType.Unknown, "", "Can't read file")
		}

		// Detect content type
		contentType := http.DetectContentType(buffer)
		ext := path.Ext(r.URL.Path)
		if contentType == "application/octet-stream" || contentType == "text/plain; charset=utf-8" {
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
		}

		// Set headers
		rw.Header().Add("Content-Type", contentType)
		rw.Header().Add("Content-Length", fmt.Sprintf("%d", totalSize))

		// Stream file
		rw.Write(buffer[0:totalSize])
		return
	} else {
		Fatal(404, ErrorType.NotFound, "", "File not found")
	}
}

func FileHandler(rw http.ResponseWriter, r *http.Request, folderPath string, url string) {
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

	// Set path
	p := url
	p = strings.ReplaceAll(p, "..", "")

	// Check file and return if found
	file := getFile(strings.ReplaceAll(folderPath+p, "//", "/"))

	if file != nil {
		stat, _ := file.Stat()

		// Check content type
		contentType, err := GetMime(folderPath + p)
		if err != nil {
			Fatal(500, ErrorType.Unknown, "", "Error while get content type")
		}

		// Set headers
		rw.Header().Add("Content-Type", contentType)
		rw.Header().Add("Content-Length", fmt.Sprintf("%d", stat.Size()))

		// Stream file
		file.Seek(0, 0)
		io.Copy(rw, file)
		return
	} else {
		Fatal(404, ErrorType.NotFound, "", "File not found")
	}
}

func ApiHandler(rw http.ResponseWriter, r *http.Request, prefix string, controller map[string]interface{}) {
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
			files := make([][]byte, 0)
			for _, fileHeaders := range r.MultipartForm.File {
				for _, fileHeader := range fileHeaders {
					f, _ := fileHeader.Open()
					defer f.Close()
					buffer := make([]byte, fileHeader.Size)
					f.Read(buffer)
					files = append(files, buffer)
				}
			}
			args["files"] = files
		}
	} else {
		// Parse json body and collect args
		jsonMap := make(map[string]interface{})
		json.NewDecoder(r.Body).Decode(&jsonMap)
		for key, element := range jsonMap {
			args[key] = element
		}
	}

	// Get controller
	path := strings.Split(strings.Replace(r.URL.Path, prefix, "", 1), "/")
	controllerName := path[1]
	methodName := ""

	if len(path) > 2 {
		methodName = path[2]
	}

	if methodName == "" {
		methodName = "Index"
	}

	// Check controller
	if controller[controllerName] == nil {
		Fatal(404, ErrorType.NotFound, "", "Controller not found")
	}

	// Call method
	context := new(RestServerContext)
	context.ContentType = "application/json"
	context.StatusCode = 200

	response, err := CallMethod(controller[controllerName], strings.Title(strings.ToLower(r.Method))+strings.Title(methodName), args, context)

	// Response is file
	if fmt.Sprintf("%v", response.Type()) == "*os.File" {
		file := response.Interface().(*os.File)
		contentType, _ := GetMimeByFile(file)
		rw.Header().Add("Content-Type", contentType)

		// Stream file
		file.Seek(0, 0)
		// io.Copy(rw, file)
		http.ServeContent(rw, r, "", time.Now(), file)
		return
	}

	if err != nil {
		Fatal(500, ErrorType.Unknown, "", err.Error())
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
}

/*func ReturnFile(path string) {
	file := getFile(path)
	if file == nil {
		Error(404, ErrorType.NotFound, "", "File not found")
	} else {
		stat, _ := file.Stat()

		// Get file header
		file.Seek(0, 0)
		buffer := make([]byte, 512)
		_, err := file.Read(buffer)
		if err != nil {
			Error(500, ErrorType.Unknown, "", "Can't read file")
		}

		// Detect content type
		contentType := http.DetectContentType(buffer)

		// Set headers
		rw.Header().Add("Content-Type", contentType)
		rw.Header().Add("Content-Length", fmt.Sprintf("%d", stat.Size()))

		// Stream file
		file.Seek(0, 0)
		io.Copy(rw, file)
	}
}*/

func Start(addr string, routers map[string]interface{}) {
	fmt.Printf("Starting server at " + addr + "\n")

	dir, err := os.Getwd()
	if err != nil {
		panic("Can't get cwd")
	}

	routers["/__debug"] = map[string]interface{}{
		"api": new(DocApi),
	}

	DocRouter = routers

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		var route interface{}
		most := ""
		prefix := ""

		for k, v := range routers {
			if strings.HasPrefix(r.URL.Path, k) {
				// fmt.Println(k, v)
				if len(most) < len(k) {
					prefix = k
					most = k
					route = v
				}
			}
		}

		if route == nil {
			Fatal(404, ErrorType.NotFound, "", "Route not found")
		}

		// Set virtual fs
		if reflect.TypeOf(route).Kind() == reflect.Struct {
			VirtualFileHandler(rw, r, route.(VirtualFs))
			return
		}

		// Set file handler
		if reflect.TypeOf(route).Kind() == reflect.String {
			folderPath := strings.ReplaceAll(strings.ReplaceAll(dir+"/"+route.(string), "\\", "/"), "//", "/")
			url := strings.Replace(r.URL.Path, prefix, "", 1)
			FileHandler(rw, r, folderPath, url)
			return
		}

		// Set api handler
		if reflect.TypeOf(route).Kind() == reflect.Map {
			controller := route.(map[string]interface{})
			ApiHandler(rw, r, most, controller)
			return
		}
	})

	// Start server
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
		return
	}
}
