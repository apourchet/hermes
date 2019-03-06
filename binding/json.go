package binding

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin/binding"
)

type JSONBinding struct{}

func (_ *JSONBinding) Bind(req *http.Request, obj interface{}) error {
	if req != nil && req.ContentLength > 0 {
		return binding.JSON.Bind(req, obj)
	}
	return nil
}

func (_ *JSONBinding) Apply(obj interface{}, req *http.Request) error {
	content, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer(content))
	req.ContentLength = int64(len(content))
	return nil
}
