package restserver

import (
	_ "embed"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type DocApi int

type XArgs struct {
	Context *RestServerContext
	X       int
	Y       string
}

type DocMethodStruct struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type DocMethod struct {
	FullPath string            `json:"fullPath"`
	Path     string            `json:"path"`
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Struct   []DocMethodStruct `json:"struct"`
}

//go:embed template/index.html
var DocMainTemplate string

var DocRouter map[string]interface{}

func (d DocApi) GetMethodList(args XArgs) []DocMethod {
	if DocRouter == nil {
		return nil
	}

	out := make([]DocMethod, 0)
	var re = regexp.MustCompile(`^(Get|Post|Patch|Delete|Put)(.*?)$`)

	for k, v := range DocRouter {
		if reflect.TypeOf(v).Kind() == reflect.Map {
			for kk, vv := range v.(map[string]interface{}) {
				fooType := reflect.TypeOf(vv)

				for i := 0; i < fooType.NumMethod(); i++ {
					method := fooType.Method(i)
					methodName := lowerFirst(re.ReplaceAllString(method.Name, "$2"))
					methodType := strings.ToUpper(re.ReplaceAllString(method.Name, `$1`))
					methodStruct := make([]DocMethodStruct, 0)

					functionType := reflect.TypeOf(method.Func.Interface())
					for i := 0; i < functionType.NumIn(); i++ {
						// Skip first argument
						if i == 0 {
							continue
						}

						argument := functionType.In(i)
						argsx := reflect.New(argument).Interface()

						s := reflect.ValueOf(argsx).Elem()
						ss := reflect.TypeOf(argsx).Elem()

						if s.Type().Kind() == reflect.Struct {
							// Go over fields
							amount := s.NumField()
							for i := 0; i < amount; i++ {
								if ss.Field(i).Name == "Context" || ss.Field(i).Name == "AccessToken" {
									continue
								}
								methodStruct = append(methodStruct, DocMethodStruct{
									Name: lowerFirst(ss.Field(i).Name),
									Type: fmt.Sprintf("%v", s.Field(i).Type()),
								})
							}
						}
					}

					out = append(out, DocMethod{
						FullPath: k + "/" + kk + "/" + methodName,
						Path:     k + "/" + kk,
						Type:     methodType,
						Name:     methodName,
						Struct:   methodStruct,
					})
				}
			}
		}
	}

	return out
}

func (d DocApi) GetIndex(args XArgs) string {
	args.Context.ContentType = "text/html"

	return DocMainTemplate
}
