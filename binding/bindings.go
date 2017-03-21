package binding

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Binding interface {
	Bind(ctx *gin.Context, obj interface{}) error
	Apply(req *http.Request, obj interface{}) error
}

type BindingFactory func(string) Binding

var DefaultBindingFactory BindingFactory = JSONBindingFactory
