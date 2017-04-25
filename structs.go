package hermes

// Input types. These interfaces must be satisfied by the user
type ICallable interface {
	SNI() string
	IServer
}

type IServer interface {
	Endpoints() EndpointMap
}

// Aliases
type EndpointMap []*Endpoint

var EP = NewEndpoint
