package binding

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var binding1 = &URLBinding{[]string{"param1", "param2"}, []string{"query1", "query2"}}

func TestTransformURLParam(t *testing.T) {
	input := struct{ Param1 string }{"myname"}
	req, _ := http.NewRequest("GET", "http://example.com/api/v1/test/:param1/something", nil)
	err := binding1.Apply(req, input)
	assert.Nil(t, err)
	assert.Equal(t, "http://example.com/api/v1/test/myname/something", req.URL.String())
}

func TestTransformURLParams(t *testing.T) {
	input := struct {
		Param1 string
		Param2 int
	}{"myname", 2}
	req, _ := http.NewRequest("GET", "http://example.com/api/v1/:param1/:param2", nil)
	err := binding1.Apply(req, input)
	assert.Nil(t, err)
	assert.Equal(t, "http://example.com/api/v1/myname/2", req.URL.String())
}

func TestTransformURLParamNil(t *testing.T) {
	input := struct {
		Param1 *string
		Param2 *int
	}{}
	p1 := "myname"
	input.Param1 = &p1
	req, _ := http.NewRequest("GET", "http://example.com/api/v1/:param1/:param2", nil)
	err := binding1.Apply(req, input)
	assert.NotNil(t, err)
}

func TestTransformURLQuery(t *testing.T) {
	input := struct{ Query1 string }{"myname"}
	req, _ := http.NewRequest("GET", "http://example.com/api/v1/test", nil)
	err := binding1.Apply(req, input)
	assert.Nil(t, err)
	assert.Equal(t, "http://example.com/api/v1/test?query1=myname", req.URL.String())
}

func TestTransformURLQueryNil(t *testing.T) {
	input := struct{ Query1 *string }{}
	req, _ := http.NewRequest("GET", "http://example.com/api/v1/test", nil)
	err := binding1.Apply(req, input)
	assert.Nil(t, err)
	assert.Equal(t, "http://example.com/api/v1/test", req.URL.String())
}

func TestTransformURLQueries(t *testing.T) {
	input := struct {
		Query1 string
		Query2 int
	}{"myname", 2}
	req, _ := http.NewRequest("GET", "http://example.com/api/v1/test", nil)
	err := binding1.Apply(req, input)
	assert.Nil(t, err)
	assert.Equal(t, "http://example.com/api/v1/test?query1=myname&query2=2", req.URL.String())
}

func TestTransformURLMix(t *testing.T) {
	input := struct {
		Param1 string
		Query2 int
	}{"myname", 2}
	req, _ := http.NewRequest("GET", "http://example.com/api/v1/:param1", nil)
	err := binding1.Apply(req, input)
	assert.Nil(t, err)
	assert.Equal(t, "http://example.com/api/v1/myname?query2=2", req.URL.String())
}
