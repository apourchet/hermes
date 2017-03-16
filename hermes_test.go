package hermes

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

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
	return http.StatusBadRequest, fmt.Errorf("Secret was wrong: '%s'", in.Message)
}

// Tests
var engine *gin.Engine

func TestMain(m *testing.M) {
	engine = gin.New()
	DefaultClient = &MockClient{"http", engine}
	NewService(&MyService{}).Serve(engine)
	os.Exit(m.Run())
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
	assert.NotNil(t, err)
}

func TestCallNotFound(t *testing.T) {
	si := NewService(&MyService{})
	err := si.Call(context.Background(), "NotAnEndpoint", &Inbound{}, &Outbound{})
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
