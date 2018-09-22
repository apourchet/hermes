package binding

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Bindable interface {
	Bind(*gin.Context) error
}

type Applyable interface {
	Apply(*http.Request) error
}

type PluginBinding struct{}

func (_ PluginBinding) Bind(ctx *gin.Context, obj interface{}) error {
	if casted, ok := obj.(Bindable); ok {
		return casted.Bind(ctx)
	}
	return nil
}

func (_ PluginBinding) Apply(req *http.Request, obj interface{}) error {
	if casted, ok := obj.(Applyable); ok {
		return casted.Apply(req)
	}
	return nil
}
