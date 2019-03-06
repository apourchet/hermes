package hermes

import (
	"net/http"
)

type HealthChecker struct{}

func (svc HealthChecker) Healthz(*http.Request) (int, error) { return http.StatusOK, nil }

var Healthz = NewEndpoint("Healthz", "GET", "/hermes/healthz", nil, nil)
