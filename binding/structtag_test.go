package binding

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type Input struct {
	AuthHeader  string        `hermes:"header=Authorization"`
	PathParam   int           `hermes:"path=someinteger"`
	QueryParam  float64       `hermes:"query=somefloat"`
	CookieParam []interface{} `hermes:"cookie=somecookie"`
}

var binding2 = &StructTagBinding{}

func TestTagBinding(t *testing.T) {
	input := &Input{"headerval", 12, 0.0, []interface{}{"a", "b"}}
	inurl := "http://example.com/:someinteger"
	req, _ := http.NewRequest("GET", inurl, nil)

	err := binding2.Apply(req, input)
	assert.Nil(t, err)

	// Check url was set properly
	assert.Equal(t, "http://example.com/12?somefloat=0", req.URL.String())

	// Check that cookie was set properly
	cookieval, err := req.Cookie("somecookie")
	assert.NoError(t, err)
	assert.Equal(t, url.PathEscape(`["a","b"]`), cookieval.Value)

	// Check that header was set properly
	val := req.Header.Get("Authorization")
	assert.Equal(t, "headerval", val)

	// Check that binding will result in the same struct
	ctx := &gin.Context{
		Request: req,
		Params: []gin.Param{
			{"someinteger", "12"},
		},
	}
	newinput := &Input{}
	err = binding2.Bind(ctx, newinput)
	assert.NoError(t, err)
	assert.Equal(t, *input, *newinput)
}
