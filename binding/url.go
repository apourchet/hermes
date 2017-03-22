package binding

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/fatih/structs"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

type URLBinding struct {
	Params  []string
	Queries []string
}

const (
	IGNORE_MULTIPLE_QUERYVALS = 1
)

var flags = 0 | IGNORE_MULTIPLE_QUERYVALS

func (b *URLBinding) Bind(ctx *gin.Context, obj interface{}) error {
	for _, param := range b.Params {
		val := ctx.Param(param)
		err := SetField(obj, param, val)
		if err != nil {
			return fmt.Errorf("Failed to set url parameter binding %s: %v", param, err)
		}
	}

	for _, query := range b.Queries {
		vals, found := ctx.Request.URL.Query()[query]
		if found && len(vals) > 0 {
			if len(vals) > 1 {
				err := fmt.Errorf("Query parameter had multiple values; which is unsupported.")
				if (flags | IGNORE_MULTIPLE_QUERYVALS) == 0 {
					glog.Warningf("%v", err)
				} else {
					return err
				}
			}
			err := SetField(obj, query, vals[0])
			if err != nil {
				return fmt.Errorf("Failed to set url query binding %s: %v", query, err)
			}
		}
	}
	return nil
}

func (b *URLBinding) Apply(req *http.Request, obj interface{}) error {
	if req == nil {
		return nil
	}

	transformedURL, err := b.TransformURL(req.URL.String(), obj)
	if err != nil {
		return fmt.Errorf("Failed to transform path with binding: %v", err)
	}

	newurl, err := url.Parse(transformedURL)
	if err != nil {
		return fmt.Errorf("Failed to parse the transformed path with binding: %v", err)
	}

	req.URL = newurl
	return nil
}

func (b *URLBinding) TransformURL(path string, input interface{}) (string, error) {
	if input == nil {
		return path, nil
	}

	v := reflect.ValueOf(input)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() { // If the chain ends in a nil, skip this
			return path, nil
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return path, nil
	}

	rawFields := structs.Map(input)
	fields := map[string]string{}
	for name, value := range rawFields {
		skip, value, err := Stringify(value)
		if err != nil {
			return "", fmt.Errorf("Failed to construct path: %v", err)
		} else if !skip {
			fields[strings.ToLower(name)] = value
		}
	}

	for _, param := range b.Params {
		if value, ok := fields[param]; ok {
			path = strings.Replace(path, ":"+param, url.QueryEscape(value), -1)
		} else if strings.Index(path, ":"+param) != -1 {
			return "", fmt.Errorf("Failed to find path parameter :%s in input %v", param, input)
		}
	}

	path += "?"
	for _, query := range b.Queries {
		if value, ok := fields[query]; ok {
			path += fmt.Sprintf("%s=%s&", url.QueryEscape(query), url.QueryEscape(value))
		}
	}

	path = strings.TrimRight(path, "&?")
	return path, nil
}

func Stringify(val interface{}) (bool, string, error) {
	v := reflect.ValueOf(val)

	// Path down the chain of pointers until a non-pointer kind
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() { // If the chain ends in a nil, skip this
			return true, "", nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Bool:
		return false, fmt.Sprintf("%v", v.Bool()), nil
	case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64:
		return false, fmt.Sprintf("%v", v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint32, reflect.Uint64:
		return false, fmt.Sprintf("%v", v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return false, fmt.Sprintf("%v", v.Float()), nil
	case reflect.String:
		return false, fmt.Sprintf("%v", v.String()), nil
	case reflect.Slice, reflect.Map:
		content, err := json.Marshal(val)
		if err != nil {
			return false, "", fmt.Errorf("Failed to stringify value into url: %v", err)
		}
		return false, string(content), nil
	}
	return false, "", fmt.Errorf("Unsupported type: %T", val)
}

// Sets the field of the object using a string that
// was retrieved from the URI of the request
func SetField(obj interface{}, fieldname, value string) error {
	v := reflect.ValueOf(obj)

	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	field := v.FieldByNameFunc(func(a string) bool { return strings.ToLower(a) == fieldname })

	if !field.IsValid() {
		return fmt.Errorf("Field not found when binding: %s", fieldname)
	}

	val, err := ParseString(field.Type(), value)
	if err != nil {
		return fmt.Errorf("Failed to parse value: %v", err)
	}

	field.Set(val)
	return nil
}

func ParseString(t reflect.Type, value string) (reflect.Value, error) {
	switch t.Kind() {
	case reflect.Ptr:
		subval, err := ParseString(t.Elem(), value)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse url parameter/query to %s: %v", t, err)
		}
		val := reflect.New(t.Elem())
		val.Elem().Set(subval)
		return val, nil
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse url parameter/query to %T: %v", b, err)
		}
		return reflect.ValueOf(b), nil
	case reflect.String:
		return reflect.ValueOf(value), nil
	case reflect.Int:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse url parameter/query to %T: %v", i, err)
		}
		return reflect.ValueOf(int(i)), nil
	case reflect.Int8:
		i, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse url parameter/query to %T: %v", i, err)
		}
		return reflect.ValueOf(int8(i)), nil
	case reflect.Int32:
		i, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse url parameter/query to %T: %v", i, err)
		}
		return reflect.ValueOf(int32(i)), nil
	case reflect.Int64:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse url parameter/query to %T: %v", i, err)
		}
		return reflect.ValueOf(int64(i)), nil
	case reflect.Uint:
		i, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse url parameter/query to %T: %v", i, err)
		}
		return reflect.ValueOf(uint(i)), nil
	case reflect.Uint8:
		i, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse url parameter/query to %T: %v", i, err)
		}
		return reflect.ValueOf(uint8(i)), nil
	case reflect.Uint32:
		i, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse url parameter/query to %T: %v", i, err)
		}
		return reflect.ValueOf(uint32(i)), nil
	case reflect.Uint64:
		i, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse url parameter/query to %T: %v", i, err)
		}
		return reflect.ValueOf(uint64(i)), nil
	case reflect.Float32:
		f, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse url parameter/query to %T: %v", f, err)
		}
		return reflect.ValueOf(float32(f)), nil
	case reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse url parameter/query to %T: %v", f, err)
		}
		return reflect.ValueOf(float64(f)), nil
	case reflect.Slice:
		var s []interface{}
		err := json.Unmarshal([]byte(value), &s)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse url parameter/query to %T: %v", s, err)
		}
		return reflect.ValueOf(s), nil
	case reflect.Map:
		var m map[string]interface{}
		err := json.Unmarshal([]byte(value), &m)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse url parameter/query to %T: %v", m, err)
		}
		return reflect.ValueOf(m), nil
	}
	return reflect.ValueOf(nil), fmt.Errorf("Unsupported type: %v", t)
}
