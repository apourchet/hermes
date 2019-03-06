package binding

import (
	"fmt"
	"net/http"
	"reflect"
)

type HeaderBinding struct {
	Headers map[string]string
}

func (b *HeaderBinding) Bind(req *http.Request, obj interface{}) error {
	for headerkey, field := range b.Headers {
		if err := BindHeader(req, obj, headerkey, field); err != nil {
			return err
		}
	}
	return nil
}

func (b *HeaderBinding) Apply(obj interface{}, req *http.Request) error {
	if req == nil || obj == nil {
		return nil
	}

	v, valid := Deref(obj)
	if !valid || v.Kind() != reflect.Struct {
		return nil
	}

	fields, err := FieldMap(obj)
	if err != nil {
		return fmt.Errorf("Failed to apply header binding: %v", err)
	}

	for headerKey, field := range b.Headers {
		if value, ok := fields[field]; ok {
			req.Header.Set(headerKey, value)
		}
	}
	return nil
}
