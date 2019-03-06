package binding

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

type URLBinding struct {
	Params  []string
	Queries []string
}

const (
	IGNORE_MULTIPLE_QUERYVALS = 1
)

var QueryFlags = 0 | IGNORE_MULTIPLE_QUERYVALS

func (b *URLBinding) Bind(req *http.Request, obj interface{}) error {
	for _, param := range b.Params {
		if err := BindPath(req, obj, param, param); err != nil {
			return err
		}
	}

	for _, query := range b.Queries {
		if err := BindQuery(req, obj, query, query); err != nil {
			return err
		}
	}
	return nil
}

func (b *URLBinding) Apply(input interface{}, req *http.Request) error {
	if req == nil || input == nil {
		return nil
	}

	v, valid := Deref(input)
	if !valid || v.Kind() != reflect.Struct {
		return nil
	}

	fields, err := FieldMap(input)
	if err != nil {
		return fmt.Errorf("Failed to apply url binding: %v", err)
	}

	for _, param := range b.Params {
		if value, ok := fields[param]; ok {
			ApplyPath(req, param, value)
		} else if strings.Index(req.URL.Path, "{"+param+"}") != -1 {
			return fmt.Errorf("Failed to find path parameter :%s in input %v", param, input)
		}
	}

	for _, query := range b.Queries {
		if value, ok := fields[query]; ok {
			ApplyQuery(req, query, value)
		}
	}

	return nil
}
