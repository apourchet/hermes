package hermes_test

import (
	"net/http"
	"testing"

	"github.com/apourchet/hermes"
	"github.com/stretchr/testify/assert"
)

type Service struct {
	hermes.HealthChecker
}

func (_ Service) SNI() string { return "UNUSED" }

func (_ Service) Endpoints() hermes.EndpointMap {
	return hermes.EndpointMap{
		hermes.Healthz,
	}
}

func TestHealthzSuccess(t *testing.T) {
	engine := http.NewServeMux()
	si := hermes.NewRouter(Service{})
	si.Serve(engine)

	caller := hermes.NewCaller(Service{})
	caller.Client = &hermes.MockClient{engine}

	code, err := caller.Call(nil, "Healthz", nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, code)
}
