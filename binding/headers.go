package binding

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/fatih/structs"
	"github.com/gin-gonic/gin"
)

type HeaderBinding struct {
	Headers map[string]string
}

func (b *HeaderBinding) Bind(ctx *gin.Context, obj interface{}) error {
	for headerKey, field := range b.Headers {
		val := ctx.Request.Header.Get(headerKey)
		if val == "" {
			continue
		}

		err := SetField(obj, field, val)
		if err != nil {
			return fmt.Errorf("Failed to set header binding %s: %v", field, err)
		}
	}
	return nil
}

func (b *HeaderBinding) Apply(req *http.Request, obj interface{}) error {
	if req == nil || obj == nil {
		return nil
	}

	v := reflect.ValueOf(obj)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() { // If the chain ends in a nil, skip this
			return nil
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	rawFields := structs.Map(obj)
	fields := map[string]string{}
	for name, value := range rawFields {
		skip, value, err := Stringify(value)
		if err != nil {
			return fmt.Errorf("Failed to construct path: %v", err)
		} else if !skip {
			fields[strings.ToLower(name)] = value
		}
	}

	for headerKey, field := range b.Headers {
		if value, ok := fields[field]; ok {
			req.Header.Set(headerKey, value)
		}
	}
	return nil
}
