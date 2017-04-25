package hermes

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/apourchet/hermes/binding"
	"github.com/gin-gonic/gin"
)

const (
	HERMES_CODE_BYPASS = -1
)

func findEndpointByHandler(svc Server, name string) (*Endpoint, error) {
	for _, ep := range svc.Endpoints() {
		if ep.Handler == name {
			return ep, nil
		}
	}
	return nil, fmt.Errorf("MethodNotFoundError")
}

func getGinHandler(svc Server, binder binding.Binding, ep *Endpoint, method reflect.Method) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Make sure there exists a request id
		EnsureRequestID(ctx)

		// Prepare inputs and outputs
		var input reflect.Value
		if ep.InputType != nil {
			input = reflect.New(ep.InputType)
		}

		var output reflect.Value
		if ep.OutputType != nil {
			output = reflect.New(ep.OutputType)
		}

		// Bind input to context
		if input.IsValid() {
			err := binder.Bind(ctx, input.Interface())
			if err != nil {
				ctx.JSON(http.StatusBadRequest, &Error{err.Error()})
				return
			}
		}

		// Prepare arguments to function
		args := []reflect.Value{reflect.ValueOf(svc), reflect.ValueOf(ctx)}
		if input.IsValid() {
			args = append(args, input)
		}
		if output.IsValid() {
			args = append(args, output)
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
			DefaultErrorHandler(ctx, ctx.Request.URL.Path, code, errVal)
			ctx.JSON(code, &Error{errVal.Error()})
		} else if output.IsValid() {
			ctx.JSON(code, output.Interface())
		} else {
			ctx.Writer.WriteHeader(code)
		}
	}
}
