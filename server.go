package hermes

import (
	"fmt"
	"reflect"

	"github.com/gin-gonic/gin"
)

// Struct wrappers
type Servable struct {
	Server   IServer
	Bindings BindingFactory
}

func NewServable(server IServer, bindings BindingFactory) Servable {
	return Servable{server, bindings}
}

func (servable Servable) Serve(engine *gin.Engine) error {
	server := servable.Server
	serviceType := reflect.TypeOf(server)
	for _, ep := range server.Endpoints() {
		method, ok := serviceType.MethodByName(ep.Handler)
		if !ok {
			return fmt.Errorf("Endpoint '%s' does not match any method of the type %v", ep.Handler, serviceType)
		}
		binding := servable.Bindings(ep.Params, ep.Queries, ep.Headers)
		fn := getGinHandler(server, binding, ep, method)
		engine.Handle(ep.Method, ep.Path, fn)
	}
	return nil
}
