package restserver

import "reflect"

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
