package restserver

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"
	"unicode"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func getFile(path string) *os.File {
	if path[len(path)-1] == '/' {
		path += "index.html"
	}

	if fileExists(path) {
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		// defer f.Close()
		return f
	}
	return nil
}

func lowerFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}

func ApplyString(field *reflect.Value, v interface{}) {
	switch reflect.TypeOf(v).Kind() {
	case reflect.String:
		field.SetString(v.(string))
	}
}

func ApplyInt(field *reflect.Value, v interface{}) {
	switch reflect.TypeOf(v).Kind() {
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
		field.SetInt(v.(int64))
	case reflect.Float32:
	case reflect.Float64:
		field.SetInt(int64(v.(float64)))
	case reflect.String:
		i, _ := strconv.ParseInt(v.(string), 10, 64)
		field.SetInt(i)
	}
}

func ApplyFloat(field *reflect.Value, v interface{}) {
	switch reflect.TypeOf(v).Kind() {
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
		field.SetFloat(float64(v.(int64)))
	case reflect.Float32:
	case reflect.Float64:
		field.SetFloat(v.(float64))
	case reflect.String:
		i, _ := strconv.ParseFloat(v.(string), 64)
		field.SetFloat(i)
	}
}

func ApplyBool(field *reflect.Value, v interface{}) {
	switch reflect.TypeOf(v).Kind() {
	case reflect.Bool:
		field.SetBool(v.(bool))
	case reflect.String:
		if v.(string) == "true" {
			field.SetBool(true)
		} else {
			field.SetBool(false)
		}
	}
}

func ApplySlice(field *reflect.Value, v interface{}) {
	len := reflect.ValueOf(v).Len()
	kind := reflect.ValueOf(field)

	if fmt.Sprintf("%v", kind) == "<json.RawMessage Value>" {
		stringJson, err := json.Marshal(v.(map[string]interface{}))
		if err == nil {
			field.Set(reflect.ValueOf(json.RawMessage(stringJson)))
		}
	}

	/*
		itemslice := reflect.SliceOf(field.Type())
		rv := reflect.New(itemslice)
		field.Set(rv.Elem())
	*/

	if fmt.Sprintf("%v", kind) == "<[]uint8 Value>" {
		slice := make([]byte, 0)
		for i := 0; i < len; i++ {
			slice = append(slice, reflect.ValueOf(v).Index(i).Interface().(byte))
		}
		field.Set(reflect.ValueOf(slice))
	}

	if fmt.Sprintf("%v", kind) == "<[][]uint8 Value>" {
		slice := make([][]byte, 0)
		for i := 0; i < len; i++ {
			slice = append(slice, reflect.ValueOf(v).Index(i).Interface().([]byte))
		}
		field.Set(reflect.ValueOf(slice))
	}

	if fmt.Sprintf("%v", kind) == "<[]string Value>" {
		slice := make([]string, 0)
		for i := 0; i < len; i++ {
			slice = append(slice, fmt.Sprintf("%v", reflect.ValueOf(v).Index(i)))
		}
		field.Set(reflect.ValueOf(slice))
	}
}

func ApplyTime(field *reflect.Value, v interface{}) {
	t := time.Now()

	t1, err := time.Parse("2006-01-02T15:04:05.999999999Z07:00", v.(string))
	if err == nil {
		t1 = time.Date(t1.Year(), t1.Month(), t1.Day(), t1.Hour(), t1.Minute(), t1.Second(), t.Nanosecond(), t.Location())
		field.Set(reflect.ValueOf(t1))
		return
	}

	t1, err = time.Parse("2006-01-02 15:04:05", v.(string))
	if err == nil {
		t1 = time.Date(t1.Year(), t1.Month(), t1.Day(), t1.Hour(), t1.Minute(), t1.Second(), t.Nanosecond(), t.Location())
		field.Set(reflect.ValueOf(t1))
		return
	}

	t1, err = time.Parse("2006-01-02", v.(string))
	if err == nil {
		t1 = time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, t.Location())
		field.Set(reflect.ValueOf(t1))
		return
	}
}

func ApplyPtr(field *reflect.Value, v interface{}) {
	field.Set(reflect.New(field.Type().Elem()))
	x := field.Elem()

	if reflect.TypeOf(v).Kind() == reflect.Map {
		FillFieldList(&x, field.Elem().Type(), v.(map[string]interface{}))
	}
}
