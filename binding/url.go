package binding

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
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

var QueryFlags = 0 | IGNORE_MULTIPLE_QUERYVALS

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
				if (QueryFlags | IGNORE_MULTIPLE_QUERYVALS) == 0 {
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
