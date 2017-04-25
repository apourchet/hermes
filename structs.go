package hermes

// Input types. These interfaces must be satisfied by the user
type IServiceable interface {
	IHosted
	IServer
}

type IHosted interface {
	SNI() string
}

type IServer interface {
	Endpoints() EndpointMap
}

// Aliases
type EndpointMap []*Endpoint

var EP = NewEndpoint
