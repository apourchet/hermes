package binding

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type PluggableInput struct {
	req *http.Request
}

func (in *PluggableInput) Bind(req *http.Request) error {
	in.req = req
	return nil
}

var pluginBinding = PluginBinding{}

func TestPluginBinding(t *testing.T) {
	in := &PluggableInput{}
	req := &http.Request{
		Method: "TESTMETHOD",
	}
	err := pluginBinding.Bind(req, in)
	require.NoError(t, err)
	require.Equal(t, "TESTMETHOD", in.req.Method)
}
