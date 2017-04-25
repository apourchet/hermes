package hermes

import "reflect"

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
