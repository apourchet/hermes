package hermes

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/apourchet/hermes/client"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.ReleaseMode) // suppress logs
}

// Service definition
type MyService struct{}

func (s *MyService) SNI() string {
	return "localhost:9000"
}

func (s *MyService) Endpoints() EndpointMap {
	return EndpointMap{
		&Endpoint{"RpcCall", "GET", "/rpccall", NewInbound, NewOutbound},
		&Endpoint{"RpcCall", "POST", "/rpccall", NewInbound, NewOutbound},
		&Endpoint{"NoInput", "POST", "/noinput", nil, NewOutbound},
		&Endpoint{"NoOutput", "POST", "/nooutput", NewInbound, nil},
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
	return http.StatusBadRequest, fmt.Errorf("Secret was wrong: '%s'", in.Message)
}

func (s *MyService) NoInput(c context.Context, out *Outbound) (int, error) {
	out.Ok = true
	return http.StatusOK, nil
}

func (s *MyService) NoOutput(c context.Context, in *Inbound) (int, error) {
	if in.Message == "secret" {
		return http.StatusOK, nil
	}
	return http.StatusBadRequest, fmt.Errorf("Secret was wrong: '%s'", in.Message)
}

// Tests
var engine *gin.Engine

func TestMain(m *testing.M) {
	engine = gin.New()
	client.DefaultClient = &client.MockClient{engine}
	NewService(&MyService{}).Serve(engine)
	os.Exit(m.Run())
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
	assert.NotNil(t, err)
	assert.True(t, out.Ok) // Error in the request, out was not filled in
}

func TestCallNotFound(t *testing.T) {
	si := NewService(&MyService{})
	err := si.Call(context.Background(), "NotAnEndpoint", &Inbound{}, &Outbound{})
	assert.NotNil(t, err)
}

func TestNoInput(t *testing.T) {
	si := NewService(&MyService{})
	out := &Outbound{false}
	err := si.Call(context.Background(), "NoInput", nil, out)
	assert.Nil(t, err)
	assert.True(t, out.Ok)
}

func TestNoOutput(t *testing.T) {
	si := NewService(&MyService{})
	err := si.Call(context.Background(), "NoOutput", &Inbound{"secret"}, nil)
	assert.Nil(t, err)

	err = si.Call(context.Background(), "NoOutput", &Inbound{"wrong secret"}, nil)
	assert.NotNil(t, err)
}

// func TestQueryParamsString(t *testing.T) {
// 	req, _ := http.NewRequest("GET", "http://localhost:9000/test?message=secret", nil)
// 	w := httptest.NewRecorder()
// 	engine.ServeHTTP(w, req)
// 	resp := w.Result()
// 	assert.NotNil(t, resp)
// 	content, _ := ioutil.ReadAll(resp.Body)
// 	assert.Equal(t, string(content), "hello")
// }
