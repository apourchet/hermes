package hermes

import (
	"context"
	"net/http"
)

type HealthCheckService struct{}

func (svc HealthCheckService) Endpoints() EndpointMap {
	return EndpointMap{EP("Healthz", "GET", "/hermes/healthz", nil, nil)}
}

func (svc HealthCheckService) Healthz(context.Context) (int, error) { return http.StatusOK, nil }

var DefaultHealthCheck Servable = NewServable(HealthCheckService{}, DefaultBindingFactory)
