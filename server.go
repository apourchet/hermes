package hermes

import (
	"fmt"
	"reflect"

	"github.com/gin-gonic/gin"
)

// Struct wrappers
type Server struct {
	Bindings BindingFactory

	handler IServer
}

func NewServer(iserver IServer) *Server {
	server := &Server{DefaultBindingFactory, iserver}
	return server
}

func (server *Server) Serve(engine *gin.Engine) error {
	handlerType := reflect.TypeOf(server.handler)
	for _, ep := range server.handler.Endpoints() {
		method, ok := handlerType.MethodByName(ep.Handler)
		if !ok {
			return fmt.Errorf("Endpoint '%s' does not match any method of the type %v", ep.Handler, handlerType)
		}
		binding := server.Bindings(ep.Params, ep.Queries, ep.Headers)
		fn := getGinHandler(server.handler, binding, ep, method)
		engine.Handle(ep.Method, ep.Path, fn)
	}
	return nil
}
