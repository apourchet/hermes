package hermes

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
)

func NewRouter(server Server) *Router {
	router := &Router{
		Bindings: DefaultBindingFactory,
		server:   server,
	}
	return router
}

func (router *Router) Serve(httpmux *http.ServeMux) error {
	mux := mux.NewRouter()
	handlerType := reflect.TypeOf(router.server)

	for _, ep := range router.server.Endpoints() {
		method, ok := handlerType.MethodByName(ep.Handler)
		if !ok {
			return fmt.Errorf("Endpoint '%s' does not match any method of the type %v", ep.Handler, handlerType)
		}
		binding := router.Bindings(ep.Params, ep.Queries, ep.Headers)
		handler := ep.Create(router.server, binding, method)
		mux.HandleFunc(ep.Path, handler).Methods(ep.Method)
	}

	httpmux.Handle("/", mux)
	return nil
}
