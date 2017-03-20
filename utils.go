package hermes

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/apourchet/hermes/binding"
	"github.com/gin-gonic/gin"
)

const (
	HERMES_CODE_BYPASS = 0
)

func findEndpointByHandler(svc Server, name string) (*Endpoint, error) {
	for _, ep := range svc.Endpoints() {
		if ep.Handler == name {
			return ep, nil
		}
	}
	return nil, fmt.Errorf("MethodNotFoundError")
}

func getGinHandler(svc Serviceable, binders binding.BindingFactory, ep *Endpoint, method reflect.Method) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input interface{}
		if ep.NewInput != nil {
			input = ep.NewInput()
		}

		var output interface{}
		if ep.NewOutput != nil {
			output = ep.NewOutput()
		}

		// Bind input to context
		if input != nil {
			err := binders(ctx).Bind(ctx.Request, input)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, &gin.H{"message": err.Error()})
				return
			}
		}

		// Prepare arguments to function
		args := []reflect.Value{reflect.ValueOf(svc), reflect.ValueOf(ctx)}
		if input != nil {
			args = append(args, reflect.ValueOf(input))
		}
		if output != nil {
			args = append(args, reflect.ValueOf(output))
		}

		// Call function
		vals := method.Func.Call(args)
		code := int(vals[0].Int())
		if code == HERMES_CODE_BYPASS {
			// Bypass code, do nothing here
			return
		}

		if !vals[1].IsNil() { // If there was an error
			errVal := vals[1].Interface().(error)
			ctx.JSON(code, map[string]string{"message": errVal.Error()})
		} else if output != nil {
			ctx.JSON(code, output)
		}
	}
}
