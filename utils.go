package hermes

import (
	"fmt"
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

func getGinHandler(svc Serviceable, binding BindingFactory, ep Endpoint, method reflect.Method) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		input := ep.NewInput()
		output := ep.NewOutput()

		err := binding(ctx).Bind(ctx.Request, input)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, &gin.H{"message": err.Error()})
			return
		}

		args := []reflect.Value{reflect.ValueOf(svc), reflect.ValueOf(ctx), reflect.ValueOf(input), reflect.ValueOf(output)}
		vals := method.Func.Call(args)
		code := int(vals[0].Int())
		if !vals[1].IsNil() {
			errVal := vals[1].Interface().(error)
			ctx.JSON(code, map[string]string{"message": errVal.Error()})
		} else {
			ctx.JSON(code, output)
		}
	}
}
