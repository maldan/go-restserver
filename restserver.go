package restserver

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

//go:embed template/ws.js
var WsJs string

var WsClientList = sync.Map{}

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
	context.StatusCode = 200

	response, err := CallMethod(controller[controllerName], strings.Title(strings.ToLower(r.Method))+strings.Title(methodName), args, context)

	// Response is file
	if fmt.Sprintf("%v", response.Type()) == "*os.File" {
		file := response.Interface().(*os.File)
		defer file.Close()

		// Detect content type
		if context.ContentType == "" {
			contentType, _ := GetMimeByFile(file)
			rw.Header().Add("Content-Type", contentType)
		} else {
			rw.Header().Add("Content-Type", context.ContentType)
		}

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
	if context.ContentType == "" {
		context.ContentType = "application/json"
	}
	rw.Header().Add("Content-Type", context.ContentType)
	if context.ContentType == "application/json" {
		responseData := RestServerResponse{Status: true}
		responseData.Response = response.Interface()
		finalData, _ := json.Marshal(responseData)
		fmt.Fprintf(rw, "%+v", string(finalData))
	} else {
		fmt.Fprintf(rw, "%+v", response)
	}
}

func WsHandler(clientId string, conn *websocket.Conn, messageType int, message []byte, controller map[string]interface{}) {
	// Parse message
	var msg WsMessage
	json.Unmarshal(message, &msg)

	// Defer recover
	defer ErrorWsMessage(conn, messageType, msg.Id)

	// Parse path
	methodPath := strings.Split(msg.Method, "/")
	if len(methodPath) < 2 {
		return
	}
	methodName := methodPath[1]

	args := make(map[string]interface{})

	context := new(RestServerContext)
	json.Unmarshal(msg.Args, &args)
	args["clientId"] = clientId

	response, err := CallMethod(controller["/ws"].(map[string]interface{})[methodPath[0]], "Ws"+strings.Title(methodName), args, context)
	var realOut []byte
	if err != nil {
		// Write message back
		realOut, _ = json.Marshal(WsResponse{
			Id:       msg.Id,
			Status:   false,
			Response: err.Error(),
		})
	} else {
		// Write message back
		realOut, _ = json.Marshal(WsResponse{
			Id:       msg.Id,
			Status:   true,
			Response: response.Interface(),
		})
	}

	err = conn.WriteMessage(messageType, realOut)
	if err != nil {
		panic("Error during message writing:" + err.Error())
	}
}

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

	// Main
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

	// Websocket
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		EnableCompression: true,
	}
	http.HandleFunc("/__ws.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/javascript; charset=utf-8")
		w.Write([]byte(strings.ReplaceAll(WsJs, "%HOST%", addr)))
	})
	http.HandleFunc("/__ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()

		// Add client to list
		clientId := UID(12)
		WsClientList.Store(clientId, WsClient{Id: clientId, Connection: conn})

		// The event loop
		for {
			// Read message
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Error during message reading:", err)
				break
			}

			// Handle
			WsHandler(clientId, conn, messageType, message, routers)
		}

		// Remove client from list
		WsClientList.Delete(clientId)
	})

	// Start server
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
		return
	}
}
