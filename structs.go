package hermes

import "github.com/apourchet/hermes/endpoint"

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
type EndpointMap []*endpoint.Endpoint

var EP = endpoint.NewEndpoint
