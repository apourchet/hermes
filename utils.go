package hermes

import (
	"fmt"
	"log"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
)

func findEndpointByHandler(svc Server, name string) (Endpoint, error) {
	for _, ep := range svc.Endpoints() {
		if ep.Handler == name {
			return ep, nil
		}
	}
	return Endpoint{}, fmt.Errorf("MethodNotFoundError")
}

func serveEndpoint(e *gin.Engine, svc Serviceable, ep Endpoint, method reflect.Method) {
	fn := getGinHandler(svc, ep, method)
	e.Handle(ep.Method, ep.Path, fn)
}

func getGinHandler(svc Serviceable, ep Endpoint, method reflect.Method) func(c *gin.Context) {
	return func(c *gin.Context) {
		input := ep.NewInput()
		output := ep.NewOutput()
		err := c.BindJSON(input)
		if err != nil {
			c.JSON(http.StatusBadRequest, &gin.H{"error": err.Error()})
			log.Println(err)
			return
		}
		args := []reflect.Value{reflect.ValueOf(svc), reflect.ValueOf(c), reflect.ValueOf(input), reflect.ValueOf(output)}
		vals := method.Func.Call(args)
		code := int(vals[0].Int())
		if !vals[1].IsNil() {
			errVal := vals[1].Interface().(error)
			c.JSON(code, map[string]string{"message": errVal.Error()})
		} else {
			c.JSON(code, output)
		}
	}
}
