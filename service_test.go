package hermes_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/apourchet/hermes"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		hermes.EP("Paramed", "GET", "/paramed/{action}", Action{}, nil).Param("action"),
		hermes.EP("Queried", "GET", "/queried", Action{}, nil).Query("action"),

		hermes.EP("TaggedParams", "GET", "/tagged/{p1}", TaggedParams{}, nil),
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

func (s *MyService) RpcCall(c *http.Request, in *Inbound, out *Outbound) (int, error) {
	if in.Message == "secret" {
		out.Ok = true
		return http.StatusOK, nil
	}
	out.Ok = false
	return http.StatusBadRequest, fmt.Errorf("Secret was wrong: '%s'", in.Message)
}

func (s *MyService) NoInput(c *http.Request, out *Outbound) (int, error) {
	out.Ok = true
	return http.StatusOK, nil
}

func (s *MyService) NoOutput(c *http.Request, in *Inbound) (int, error) {
	if in.Message == "secret" {
		return http.StatusOK, nil
	}
	return http.StatusBadRequest, fmt.Errorf("Secret was wrong: '%s'", in.Message)
}

func (s *MyService) Paramed(c *http.Request, in *Action) (int, error) {
	if in.Action == 69 {
		return http.StatusOK, nil
	}
	return http.StatusBadRequest, fmt.Errorf("Action should be %d", 69)
}

func (s *MyService) Queried(c *http.Request, in *Action) (int, error) {
	if in.Action == 69 {
		return http.StatusOK, nil
	}
	return http.StatusBadRequest, fmt.Errorf("Action should be %d", 69)
}

func (s *MyService) Sliced(c *http.Request, in *Slice) (int, error) {
	if len(*in) == 2 {
		return http.StatusOK, nil
	}
	return http.StatusBadRequest, fmt.Errorf("Slice should have length 2")
}

func (s *MyService) Pointers(c *http.Request, in *Pointers) (int, error) {
	if in.I != nil && *in.I == 69 && in.S == nil {
		return http.StatusOK, nil
	}
	return http.StatusBadRequest, fmt.Errorf("Should be 69 and nil; have %v and %v", in.I, in.S)
}

func (s *MyService) QueryPointers(c *http.Request, in *Pointers) (int, error) {
	if in.I != nil && *in.I == 69 && in.S == nil {
		return http.StatusOK, nil
	}
	return http.StatusBadRequest, fmt.Errorf("Should be 69 and nil; have %v and %v", in.I, in.S)
}

func (s *MyService) AllTypes(c *http.Request, in *AllTypes) (int, error) {
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

func (s *MyService) RawType(c *http.Request, in *int, out *string) (int, error) {
	str := fmt.Sprintf("%d", *in)
	*out = str
	return http.StatusOK, nil
}

func (s *MyService) RawMap(c *http.Request, in *map[string]string, out *map[string]int) (int, error) {
	*out = map[string]int{}
	for k, v := range *in {
		(*out)[k+v] = 13
	}
	return http.StatusOK, nil
}

func (s *MyService) TaggedParams(c *http.Request, in *TaggedParams) (int, error) {
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
var si *hermes.Caller
var engine *http.ServeMux

func TestMain(m *testing.M) {
	engine = http.NewServeMux()
	server := hermes.NewRouter(&MyService{})
	server.Serve(engine)

	caller := hermes.NewCaller(&MyService{})
	caller.Client = &hermes.MockClient{engine}
	si = caller
	os.Exit(m.Run())
}

func TestCallSuccess(t *testing.T) {
	out := &Outbound{false}
	_, err := si.Call(nil, "RpcCall", &Inbound{"secret"}, out)
	require.NoError(t, err)
	assert.True(t, out.Ok)
}

func TestCallWrongSecret(t *testing.T) {
	out := &Outbound{true}
	_, err := si.Call(nil, "RpcCall", &Inbound{"wrong secret"}, out)
	assert.NotNil(t, err)
	assert.True(t, out.Ok) // Error in the request, out was not filled in
}

func TestCallNotFound(t *testing.T) {
	_, err := si.Call(nil, "NotAnEndpoint", &Inbound{}, &Outbound{})
	assert.NotNil(t, err)
}

func TestNoInput(t *testing.T) {
	out := &Outbound{false}
	_, err := si.Call(nil, "NoInput", nil, out)
	require.NoError(t, err)
	assert.True(t, out.Ok)
}

func TestNoOutput(t *testing.T) {
	_, err := si.Call(nil, "NoOutput", &Inbound{"secret"}, nil)
	require.NoError(t, err)

	_, err = si.Call(nil, "NoOutput", &Inbound{"wrong secret"}, nil)
	assert.NotNil(t, err)
}

func TestSliced(t *testing.T) {
	_, err := si.Call(nil, "Sliced", &Slice{"a", "b"}, nil)
	require.NoError(t, err)

	_, err = si.Call(nil, "Sliced", nil, nil)
	assert.NotNil(t, err)
}

func TestPointers(t *testing.T) {
	s := ""
	i := 69
	_, err := si.Call(nil, "Pointers", &Pointers{&i, &s}, nil)
	assert.NotNil(t, err)

	_, err = si.Call(nil, "Pointers", &Pointers{&i, nil}, nil)
	require.NoError(t, err)
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

	_, err := si.Call(nil, "AllTypes", input, nil)
	require.NoError(t, err)
}

func TestRawType(t *testing.T) {
	in := 13
	out := ""
	_, err := si.Call(nil, "RawType", &in, &out)
	require.NoError(t, err)
	assert.Equal(t, "13", out)
}

func TestRawMap(t *testing.T) {
	in := map[string]string{"a": "b"}
	out := map[string]int{}
	_, err := si.Call(nil, "RawMap", &in, &out)
	require.NoError(t, err)
	assert.Equal(t, 1, len(out))
	assert.Equal(t, 13, out["ab"])
}

func TestQueryPointers(t *testing.T) {
	s := ""
	i := 69
	_, err := si.Call(nil, "QueryPointers", &Pointers{&i, &s}, nil)
	assert.NotNil(t, err)

	_, err = si.Call(nil, "QueryPointers", &Pointers{&i, nil}, nil)
	require.NoError(t, err)
}

func TestParamed(t *testing.T) {
	_, err := si.Call(nil, "Paramed", &Action{69}, nil)
	require.NoError(t, err)

	_, err = si.Call(nil, "Paramed", &Action{13}, nil)
	assert.NotNil(t, err)
}

func TestQueried(t *testing.T) {
	_, err := si.Call(nil, "Queried", &Action{69}, nil)
	require.NoError(t, err)

	_, err = si.Call(nil, "Queried", &Action{13}, nil)
	assert.NotNil(t, err)
}

func TestTaggedParam(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.org/tagged/mypath?q1=42", nil)
	require.NoError(t, err)
	req.Header.Add("h1", "myheader")
	req = mux.SetURLVars(req, map[string]string{
		"p1": "mypath",
	})

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
