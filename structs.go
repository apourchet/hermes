package hermes

// Input types. These interfaces must be satisfied by the user
type Serviceable interface {
	Hosted
	Server
	// Should also implement all endpoints
}

type EndpointMap []Endpoint

type Endpoint struct {
	Handler   string
	Method    string
	Path      string
	NewInput  func() interface{}
	NewOutput func() interface{}
}

type Hosted interface {
	Host() string
}

type Server interface {
	Endpoints() EndpointMap
}
