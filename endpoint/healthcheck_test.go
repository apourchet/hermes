package endpoint_test

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/apourchet/hermes"
	"github.com/apourchet/hermes/client"
	"github.com/apourchet/hermes/endpoint"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type Service struct {
	endpoint.HealthChecker
}

func (_ Service) SNI() string { return "UNUSED" }

func (_ Service) Endpoints() hermes.EndpointMap {
	return hermes.EndpointMap{
		endpoint.Healthz,
	}
}

var engine = gin.New()
var si *hermes.Service

func TestMain(m *testing.M) {
	client.DefaultClient = &client.MockClient{engine}
	si = hermes.NewService(Service{})
	si.Serve(engine)
	os.Exit(m.Run())
}

func TestHealthzSuccess(t *testing.T) {
	code, err := si.Call(context.Background(), "Healthz", nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, code)
}
