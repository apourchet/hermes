package hermes

import "github.com/apourchet/hermes/endpoint"

// Input types. These interfaces must be satisfied by the user
type Serviceable interface {
	Hosted
	Server
}

type Hosted interface {
	SNI() string
}

type Server interface {
	Endpoints() EndpointMap
}

type EndpointMap []*endpoint.Endpoint

var EP = endpoint.NewEndpoint
