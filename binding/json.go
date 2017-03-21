package binding

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type JSONBinding struct{}

func (_ *JSONBinding) Bind(ctx *gin.Context, obj interface{}) error {
	if ctx.Request != nil && ctx.Request.ContentLength > 0 {
		return binding.JSON.Bind(ctx.Request, obj)
	}
	return nil
}

func (_ *JSONBinding) Apply(req *http.Request, obj interface{}) error {
	content, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer(content))
	req.ContentLength = int64(len(content))
	return nil
}

func JSONBindingFactory(_ string) Binding {
	return &JSONBinding{}
}
