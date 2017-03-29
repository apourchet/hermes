package endpoint

import (
	"context"
	"net/http"
)

type HealthChecker struct{}

func (svc HealthChecker) Healthz(context.Context) (int, error) { return http.StatusOK, nil }

var Healthz = NewEndpoint("Healthz", "GET", "/hermes/healthz", nil, nil)
