package binding

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Input struct {
	AuthHeader string  `hermes:"header=Authorization"`
	PathParam  int     `hermes:"path=someinteger"`
	QueryParam float64 `hermes:"query=somefloat"`
}

var binding2 = &StructTagBinding{}

func TestTagBinding(t *testing.T) {
	input := &Input{"headerval", 12, 0.0}
	url := "http://example.com/:someinteger"
	resulturl := "http://example.com/12?somefloat=0"
	req, _ := http.NewRequest("GET", url, nil)
	err := binding2.Apply(req, input)
	assert.Nil(t, err)
	assert.Equal(t, resulturl, req.URL.String())
}
