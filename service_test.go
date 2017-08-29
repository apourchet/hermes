package hermes_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/apourchet/hermes"
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
		hermes.EP("RpcCall", "GET", "/rpccall", Inbound{}, Outbound{}),
		hermes.EP("NoInput", "POST", "/noinput", nil, Outbound{}),
		hermes.EP("NoOutput", "POST", "/nooutput", Inbound{}, nil),
		hermes.EP("Sliced", "GET", "/sliced", Slice{}, nil),
		hermes.EP("Pointers", "GET", "/pointers", Pointers{}, nil),
		hermes.EP("AllTypes", "GET", "/alltypes", AllTypes{}, nil),
		hermes.EP("RawType", "GET", "/rawtype", 0, ""),
		hermes.EP("RawMap", "GET", "/rawmap", map[string]string{}, map[string]int{}),

		hermes.EP("QueryPointers", "GET", "/parampointers", Pointers{}, nil).Query("i", "s"),
		hermes.EP("Paramed", "GET", "/paramed/:action", Action{}, nil).Param("action"),
		hermes.EP("Queried", "GET", "/queried", Action{}, nil).Query("action"),

		hermes.EP("TaggedParams", "GET", "/tagged/:p1", TaggedParams{}, nil),
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
type TaggedParams struct {
	Path   string  `hermes:"path=p1"`
	Query  int     `hermes:"query=q1"`
	Header *string `hermes:"header=h1"`
}

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

func (s *MyService) RawType(c context.Context, in *int, out *string) (int, error) {
	str := fmt.Sprintf("%d", *in)
	*out = str
	return http.StatusOK, nil
}

func (s *MyService) RawMap(c context.Context, in *map[string]string, out *map[string]int) (int, error) {
	*out = map[string]int{}
	for k, v := range *in {
		(*out)[k+v] = 13
	}
	return http.StatusOK, nil
}

func (s *MyService) TaggedParams(c context.Context, in *TaggedParams) (int, error) {
	if in.Path != "mypath" {
		return http.StatusBadRequest, nil
	}
	if in.Query != 42 {
		return http.StatusBadRequest, nil
	}
	if in.Header == nil || *(in.Header) != "myheader" {
		return http.StatusBadRequest, nil
	}
	return http.StatusOK, nil
}

// Tests
var si hermes.ICaller
var engine *gin.Engine

func TestMain(m *testing.M) {
	engine = gin.New()
	server := hermes.NewRouter(&MyService{})
	server.Serve(engine)

	caller := hermes.NewCaller(&MyService{})
	caller.Client = &hermes.MockClient{engine}
	si = caller
	os.Exit(m.Run())
}

func TestCallSuccess(t *testing.T) {
	out := &Outbound{false}
	_, err := si.Call(context.Background(), "RpcCall", &Inbound{"secret"}, out)
	assert.Nil(t, err)
	assert.True(t, out.Ok)
}

func TestCallWrongSecret(t *testing.T) {
	out := &Outbound{true}
	_, err := si.Call(context.Background(), "RpcCall", &Inbound{"wrong secret"}, out)
	assert.NotNil(t, err)
	assert.True(t, out.Ok) // Error in the request, out was not filled in
}

func TestCallNotFound(t *testing.T) {
	_, err := si.Call(context.Background(), "NotAnEndpoint", &Inbound{}, &Outbound{})
	assert.NotNil(t, err)
}

func TestNoInput(t *testing.T) {
	out := &Outbound{false}
	_, err := si.Call(context.Background(), "NoInput", nil, out)
	assert.Nil(t, err)
	assert.True(t, out.Ok)
}

func TestNoOutput(t *testing.T) {
	_, err := si.Call(context.Background(), "NoOutput", &Inbound{"secret"}, nil)
	assert.Nil(t, err)

	_, err = si.Call(context.Background(), "NoOutput", &Inbound{"wrong secret"}, nil)
	assert.NotNil(t, err)
}

func TestSliced(t *testing.T) {
	_, err := si.Call(context.Background(), "Sliced", &Slice{"a", "b"}, nil)
	assert.Nil(t, err)

	_, err = si.Call(context.Background(), "Sliced", nil, nil)
	assert.NotNil(t, err)
}

func TestPointers(t *testing.T) {
	s := ""
	i := 69
	_, err := si.Call(context.Background(), "Pointers", &Pointers{&i, &s}, nil)
	assert.NotNil(t, err)

	_, err = si.Call(context.Background(), "Pointers", &Pointers{&i, nil}, nil)
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

	_, err := si.Call(context.Background(), "AllTypes", input, nil)
	assert.Nil(t, err)
}

func TestRawType(t *testing.T) {
	in := 13
	out := ""
	_, err := si.Call(context.Background(), "RawType", &in, &out)
	assert.Nil(t, err)
	assert.Equal(t, "13", out)
}

func TestRawMap(t *testing.T) {
	in := map[string]string{"a": "b"}
	out := map[string]int{}
	_, err := si.Call(context.Background(), "RawMap", &in, &out)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(out))
	assert.Equal(t, 13, out["ab"])
}

func TestQueryPointers(t *testing.T) {
	s := ""
	i := 69
	_, err := si.Call(context.Background(), "QueryPointers", &Pointers{&i, &s}, nil)
	assert.NotNil(t, err)

	_, err = si.Call(context.Background(), "QueryPointers", &Pointers{&i, nil}, nil)
	assert.Nil(t, err)
}

func TestParamed(t *testing.T) {
	_, err := si.Call(context.Background(), "Paramed", &Action{69}, nil)
	assert.Nil(t, err)

	_, err = si.Call(context.Background(), "Paramed", &Action{13}, nil)
	assert.NotNil(t, err)
}

func TestQueried(t *testing.T) {
	_, err := si.Call(context.Background(), "Queried", &Action{69}, nil)
	assert.Nil(t, err)

	_, err = si.Call(context.Background(), "Queried", &Action{13}, nil)
	assert.NotNil(t, err)
}

func TestTaggedParam(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.org/tagged/mypath?q1=42", nil)
	assert.Nil(t, err)
	req.Header.Add("h1", "myheader")

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
