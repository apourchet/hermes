package binding

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type BindingFactorySource func(string) BindingFactory
type BindingFactory func(*gin.Context) binding.Binding

var DefaultBindingFactorySource BindingFactorySource = JSONBindingFactorySource
var DefaultBindingFactory BindingFactory = JSONBindingFactory

func JSONBindingFactorySource(_ string) BindingFactory {
	return JSONBindingFactory
}

func JSONBindingFactory(_ *gin.Context) binding.Binding {
	return binding.JSON
}
