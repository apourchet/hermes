package hermes

import (
	"fmt"
	"reflect"

	"github.com/gin-gonic/gin"
)

// Struct wrappers
type Router struct {
	Bindings BindingFactory

	server Server
}

func NewRouter(server Server) *Router {
	router := &Router{DefaultBindingFactory, server}
	return router
}

func (router *Router) Serve(engine *gin.Engine) error {
	handlerType := reflect.TypeOf(router.server)
	for _, ep := range router.server.Endpoints() {
		method, ok := handlerType.MethodByName(ep.Handler)
		if !ok {
			return fmt.Errorf("Endpoint '%s' does not match any method of the type %v", ep.Handler, handlerType)
		}
		binding := router.Bindings(ep.Params, ep.Queries, ep.Headers)
		fn := getGinHandler(router.server, binding, ep, method)
		engine.Handle(ep.Method, ep.Path, fn)
	}
	return nil
}
