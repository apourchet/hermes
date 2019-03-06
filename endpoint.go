package hermes

import (
	"net/http"
	"reflect"

	"github.com/apourchet/hermes/binding"
)

type Endpoint struct {
	Handler    string
	Method     string
	Path       string
	InputType  reflect.Type
	OutputType reflect.Type

	Params  []string
	Queries []string
	Headers map[string]string
}

func NewEndpoint(handler, method, path string, input, output interface{}) *Endpoint {
	ep := &Endpoint{}
	ep.Handler = handler
	ep.Method = method
	ep.Path = path
	ep.Headers = map[string]string{}

	if input != nil {
		ep.InputType = reflect.TypeOf(input)
	}

	if output != nil {
		ep.OutputType = reflect.TypeOf(output)
	}
	return ep
}

func (ep *Endpoint) Param(varnames ...string) *Endpoint {
	ep.Params = append(ep.Params, varnames...)
	return ep
}

func (ep *Endpoint) Query(varnames ...string) *Endpoint {
	ep.Queries = append(ep.Queries, varnames...)
	return ep
}

func (ep *Endpoint) Header(varname, fieldname string) *Endpoint {
	ep.Headers[varname] = fieldname
	return ep
}

func (ep *Endpoint) Create(svc interface{}, binder binding.Binding, method reflect.Method) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Make sure there exists a request id
		EnsureRequestID(req)

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
			err := binder.Bind(req, input.Interface())
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write(Error{err.Error()}.JSON())
				return
			}
		}

		// Prepare arguments to function
		args := []reflect.Value{reflect.ValueOf(svc), reflect.ValueOf(req)}
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
			DefaultErrorHandler(req, code, errVal)
			w.WriteHeader(code)
			w.Write(Error{errVal.Error()}.JSON())
		} else if output.IsValid() {
			DefaultSuccessHandler(req, code)
			w.WriteHeader(code)
			w.Write(shouldJSON(output.Interface()))
		} else {
			DefaultSuccessHandler(req, code)
			w.WriteHeader(code)
		}
	}
}
