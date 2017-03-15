package hermes

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Service definition
type MyService struct{}

func (s *MyService) Host() string {
	return "localhost:9000"
}

func (s *MyService) Endpoints() EndpointMap {
	return EndpointMap{
		Endpoint{"RpcCall", "GET", "/test", NewInbound, NewOutbound},
		Endpoint{"RpcCall", "POST", "/test", NewInbound, NewOutbound},
	}
}

// Endpoint definitions
type Inbound struct {
	Message string
}

type Outbound struct {
	Ok bool
}

func NewInbound() interface{}  { return &Inbound{} }
func NewOutbound() interface{} { return &Outbound{} }

func (s *MyService) RpcCall(c context.Context, in *Inbound, out *Outbound) (int, error) {
	if in.Message == "secret" {
		out.Ok = true
		return http.StatusOK, nil
	}
	out.Ok = false
	return http.StatusBadRequest, fmt.Errorf("Secret was wrong")
}

// Tests
func TestMain(m *testing.M) {
	engine := gin.New()
	DefaultClient = &MockClient{"http", engine}
	NewService(&MyService{}).Serve(engine)
	m.Run()
}

func TestMock(t *testing.T) {
	si := NewMockService(&MyService{})
	out := &Outbound{false}
	err := si.Call(context.Background(), "RpcCall", &Inbound{"secret"}, out)
	assert.Nil(t, err)
	assert.True(t, out.Ok)
}

func TestCallSuccess(t *testing.T) {
	si := NewService(&MyService{})
	out := &Outbound{false}
	err := si.Call(context.Background(), "RpcCall", &Inbound{"secret"}, out)
	assert.Nil(t, err)
	assert.True(t, out.Ok)
}

func TestCallWrongSecret(t *testing.T) {
	si := NewService(&MyService{})
	out := &Outbound{true}
	err := si.Call(context.Background(), "RpcCall", &Inbound{"wrong secret"}, out)
	assert.Nil(t, err)
	assert.False(t, out.Ok)
}

func TestCallNotFound(t *testing.T) {
	si := NewService(&MyService{})
	err := si.Call(context.Background(), "NotAnEndpoint", &Inbound{}, &Outbound{})
	assert.NotNil(t, err)
}
