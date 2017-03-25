package endpoint

type Endpoint struct {
	Handler   string
	Method    string
	Path      string
	NewInput  func() interface{}
	NewOutput func() interface{}

	Params  []string
	Queries []string
	Headers map[string]string
}

func NewEndpoint(handler, method, path string, input, output func() interface{}) *Endpoint {
	ep := &Endpoint{}
	ep.Handler = handler
	ep.Method = method
	ep.Path = path
	ep.NewInput = input
	ep.NewOutput = output
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
