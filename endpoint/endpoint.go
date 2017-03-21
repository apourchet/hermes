package endpoint

type Endpoint struct {
	Handler   string
	Method    string
	Path      string
	NewInput  func() interface{}
	NewOutput func() interface{}

	Params  []string
	Queries []string
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

func (ep *Endpoint) Param(varname string) *Endpoint {
	ep.Params = append(ep.Params, varname)
	return ep
}

func (ep *Endpoint) Query(varname string) *Endpoint {
	ep.Queries = append(ep.Queries, varname)
	return ep
}
