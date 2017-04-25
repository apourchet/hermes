package hermes_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/apourchet/hermes"
	"github.com/gin-gonic/gin"
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
	engine := gin.New()
	si := hermes.NewRouter(Service{})
	si.Serve(engine)

	caller := hermes.NewCaller(Service{})
	caller.Client = &hermes.MockClient{engine}

	code, err := caller.Call(context.Background(), "Healthz", nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, code)
}
