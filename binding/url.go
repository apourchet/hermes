package binding

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/fatih/structs"
	"github.com/gin-gonic/gin"
)

type URLBinding struct {
	Params  []string
	Queries []string
}

func (_ *URLBinding) Bind(ctx *gin.Context, obj interface{}) error {
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

	rawFields := structs.Map(input)
	fields := map[string]string{}
	for name, value := range rawFields {
		value, err := Stringify(value)
		if err != nil {
			return "", fmt.Errorf("Failed to construct path: %v", err)
		}
		fields[strings.ToLower(name)] = value
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

func Stringify(val interface{}) (string, error) {
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Bool:
		return fmt.Sprintf("%v", v.Bool()), nil
	case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%v", v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%v", v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%v", v.Float()), nil
	case reflect.String:
		return fmt.Sprintf("%v", v.String()), nil
	case reflect.Slice, reflect.Map:
		content, err := json.Marshal(val)
		if err != nil {
			return "", fmt.Errorf("Failed to stringify value into url: %v", err)
		}
		return string(content), nil
	}
	return "", fmt.Errorf("Unsupported type: %T", val)
}
