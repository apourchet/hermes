package hermes

// Input types. These interfaces must be satisfied by the user
type Serviceable interface {
	Hosted
	Server
}

type EndpointMap []*Endpoint

type Endpoint struct {
	Handler   string
	Method    string
	Path      string
	NewInput  func() interface{}
	NewOutput func() interface{}
}

type Hosted interface {
	SNI() string
}

type Server interface {
	Endpoints() EndpointMap
}
