package binding

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testInput struct {
	AuthHeader  string        `hermes:"header=Authorization"`
	PathParam   int           `hermes:"path=someinteger"`
	QueryParam  float64       `hermes:"query=somefloat"`
	CookieParam []interface{} `hermes:"cookie=somecookie"`
}

func TestTagBinding(t *testing.T) {
	binding2 := StructTagBinding{}

	t.Run("test bind", func(t *testing.T) {
		inurl := "http://example.com/12?somefloat=1.1"
		req, _ := http.NewRequest("GET", inurl, nil)
		req.Header.Set("Authorization", "TOKEN")
		req.AddCookie(&http.Cookie{Name: "somecookie", Value: url.PathEscape(`["a","b"]`)})
		req = mux.SetURLVars(req, map[string]string{
			"someinteger": "12",
		})

		// Check that binding will result in the same struct
		newinput := &testInput{}
		err := binding2.Bind(req, newinput)
		require.NoError(t, err)
		assert.Equal(t, "TOKEN", newinput.AuthHeader)
		assert.Equal(t, 12, newinput.PathParam)
		assert.Equal(t, 1.1, newinput.QueryParam)
		assert.Len(t, newinput.CookieParam, 2)
	})

	t.Run("test apply", func(t *testing.T) {
		input := &testInput{"headerval", 12, 0.0, []interface{}{"a", "b"}}
		inurl := "http://example.com/{someinteger}"
		req, _ := http.NewRequest("GET", inurl, nil)

		err := binding2.Apply(input, req)
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
	})

}
