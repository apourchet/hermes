package hermes_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/apourchet/hermes"
	"github.com/apourchet/hermes/client"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.ReleaseMode) // suppress logs
}

// Service definition
type MyService struct{}

func (s *MyService) SNI() string { return "localhost:9000" }

func (s *MyService) Endpoints() hermes.EndpointMap {
	return hermes.EndpointMap{
		hermes.EP("RpcCall", "GET", "/rpccall", NewInbound, NewOutbound),
		hermes.EP("RpcCall", "POST", "/rpccall", NewInbound, NewOutbound),
		hermes.EP("NoInput", "POST", "/noinput", nil, NewOutbound),
		hermes.EP("NoOutput", "POST", "/nooutput", NewInbound, nil),
		hermes.EP("Sliced", "GET", "/sliced", NewSlice, nil),
		hermes.EP("Pointers", "GET", "/pointers", NewPointers, nil),
		hermes.EP("AllTypes", "GET", "/alltypes", NewAllTypes, nil),

		hermes.EP("QueryPointers", "GET", "/parampointers", NewPointers, nil).Query("i", "s"),
		hermes.EP("Paramed", "GET", "/paramed/:action", NewAction, nil).Param("action"),
		hermes.EP("Queried", "GET", "/queried", NewAction, nil).Query("action"),
	}
}

// Endpoint definitions
type Inbound struct{ Message string }
type Outbound struct{ Ok bool }
type Action struct{ Action int }
type Slice []string
type Pointers struct {
	I *int
	S *string
}
type AllTypes struct {
	A bool
	B int
	C string
	P *int
	M map[string]string
	S struct {
		E string
		F *float64
	}
	Sl []struct {
		E string
		F *float64
	}
}

func NewInbound() interface{}  { return &Inbound{} }
func NewOutbound() interface{} { return &Outbound{} }
func NewAction() interface{}   { return &Action{} }
func NewSlice() interface{}    { return &Slice{} }
func NewPointers() interface{} { return &Pointers{} }
func NewAllTypes() interface{} { return &AllTypes{} }

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

func (s *MyService) Paramed(c context.Context, in *Action) (int, error) {
	if in.Action == 69 {
		return http.StatusOK, nil
	}
	return http.StatusBadRequest, fmt.Errorf("Action should be %d", 69)
}

func (s *MyService) Queried(c context.Context, in *Action) (int, error) {
	if in.Action == 69 {
		return http.StatusOK, nil
	}
	return http.StatusBadRequest, fmt.Errorf("Action should be %d", 69)
}

func (s *MyService) Sliced(c context.Context, in *Slice) (int, error) {
	if len(*in) == 2 {
		return http.StatusOK, nil
	}
	return http.StatusBadRequest, fmt.Errorf("Slice should have length 2")
}

func (s *MyService) Pointers(c context.Context, in *Pointers) (int, error) {
	if in.I != nil && *in.I == 69 && in.S == nil {
		return http.StatusOK, nil
	}
	return http.StatusBadRequest, fmt.Errorf("Should be 69 and nil; have %v and %v", in.I, in.S)
}

func (s *MyService) QueryPointers(c context.Context, in *Pointers) (int, error) {
	if in.I != nil && *in.I == 69 && in.S == nil {
		return http.StatusOK, nil
	}
	return http.StatusBadRequest, fmt.Errorf("Should be 69 and nil; have %v and %v", in.I, in.S)
}

func (s *MyService) AllTypes(c context.Context, in *AllTypes) (int, error) {
	p := 1
	f := float64(3.14)
	wanted := &AllTypes{
		A: true,
		B: 69,
		C: "test",
		P: &p,
		M: map[string]string{"a": "b", "e": "f"},
		S: struct {
			E string
			F *float64
		}{},
		Sl: []struct {
			E string
			F *float64
		}{{"slice", &f}},
	}
	if reflect.DeepEqual(in, wanted) {
		return http.StatusOK, nil
	}
	return http.StatusBadRequest, fmt.Errorf("Wanted %v, got %v", wanted, in)
}

// Tests
var engine *gin.Engine
var si *hermes.Service

func TestMain(m *testing.M) {
	engine = gin.New()
	client.DefaultClient = &client.MockClient{engine}
	si = hermes.NewService(&MyService{})
	si.Serve(engine)
	os.Exit(m.Run())
}

func TestCallSuccess(t *testing.T) {
	out := &Outbound{false}
	err := si.Call(context.Background(), "RpcCall", &Inbound{"secret"}, out)
	assert.Nil(t, err)
	assert.True(t, out.Ok)
}

func TestCallWrongSecret(t *testing.T) {
	out := &Outbound{true}
	err := si.Call(context.Background(), "RpcCall", &Inbound{"wrong secret"}, out)
	assert.NotNil(t, err)
	assert.True(t, out.Ok) // Error in the request, out was not filled in
}

func TestCallNotFound(t *testing.T) {
	err := si.Call(context.Background(), "NotAnEndpoint", &Inbound{}, &Outbound{})
	assert.NotNil(t, err)
}

func TestNoInput(t *testing.T) {
	out := &Outbound{false}
	err := si.Call(context.Background(), "NoInput", nil, out)
	assert.Nil(t, err)
	assert.True(t, out.Ok)
}

func TestNoOutput(t *testing.T) {
	err := si.Call(context.Background(), "NoOutput", &Inbound{"secret"}, nil)
	assert.Nil(t, err)

	err = si.Call(context.Background(), "NoOutput", &Inbound{"wrong secret"}, nil)
	assert.NotNil(t, err)
}

func TestSliced(t *testing.T) {
	err := si.Call(context.Background(), "Sliced", &Slice{"a", "b"}, nil)
	assert.Nil(t, err)

	err = si.Call(context.Background(), "Sliced", nil, nil)
	assert.NotNil(t, err)
}

func TestPointers(t *testing.T) {
	s := ""
	i := 69
	err := si.Call(context.Background(), "Pointers", &Pointers{&i, &s}, nil)
	assert.NotNil(t, err)

	err = si.Call(context.Background(), "Pointers", &Pointers{&i, nil}, nil)
	assert.Nil(t, err)
}

func TestAllTypes(t *testing.T) {
	p := 1
	f := float64(3.14)
	input := &AllTypes{
		A: true,
		B: 69,
		C: "test",
		P: &p,
		M: map[string]string{"a": "b", "e": "f"},
		S: struct {
			E string
			F *float64
		}{},
		Sl: []struct {
			E string
			F *float64
		}{{"slice", &f}},
	}

	err := si.Call(context.Background(), "AllTypes", input, nil)
	assert.Nil(t, err)
}

func TestQueryPointers(t *testing.T) {
	s := ""
	i := 69
	err := si.Call(context.Background(), "QueryPointers", &Pointers{&i, &s}, nil)
	assert.NotNil(t, err)

	err = si.Call(context.Background(), "QueryPointers", &Pointers{&i, nil}, nil)
	assert.Nil(t, err)
}

func TestParamed(t *testing.T) {
	err := si.Call(context.Background(), "Paramed", &Action{69}, nil)
	assert.Nil(t, err)

	err = si.Call(context.Background(), "Paramed", &Action{13}, nil)
	assert.NotNil(t, err)
}

func TestQueried(t *testing.T) {
	err := si.Call(context.Background(), "Queried", &Action{69}, nil)
	assert.Nil(t, err)

	err = si.Call(context.Background(), "Queried", &Action{13}, nil)
	assert.NotNil(t, err)
}
