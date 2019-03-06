package binding

import (
	"net/http"
)

type Bindable interface {
	Bind(*http.Request) error
}

type Applyable interface {
	Apply(*http.Request) error
}

type PluginBinding struct{}

func (_ PluginBinding) Bind(req *http.Request, obj interface{}) error {
	if casted, ok := obj.(Bindable); ok {
		return casted.Bind(req)
	}
	return nil
}

func (_ PluginBinding) Apply(obj interface{}, req *http.Request) error {
	if casted, ok := obj.(Applyable); ok {
		return casted.Apply(req)
	}
	return nil
}
