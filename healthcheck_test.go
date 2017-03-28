package hermes_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/apourchet/hermes"
	"github.com/apourchet/hermes/client"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthz(t *testing.T) {
	engine := gin.New()
	client.DefaultClient = &client.MockClient{engine}
	hermes.DefaultHealthCheck.Serve(engine)

	req, err := http.NewRequest("GET", "/hermes/healthz", nil)
	assert.Nil(t, err)
	client.DefaultClient.Exec(context.Background(), req)
}
