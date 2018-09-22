package binding

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type PluggableInput struct {
	req *http.Request
}

func (in *PluggableInput) Bind(ctx *gin.Context) error {
	in.req = ctx.Request
	return nil
}

var pluginBinding = PluginBinding{}

func TestPluginBinding(t *testing.T) {
	in := &PluggableInput{}
	ctx := &gin.Context{
		Request: &http.Request{
			Method: "TESTMETHOD",
		},
	}
	err := pluginBinding.Bind(ctx, in)
	require.NoError(t, err)
	require.Equal(t, "TESTMETHOD", in.req.Method)
}
