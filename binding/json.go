package binding

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo"
)

type JSONBinding struct{}

func (_ *JSONBinding) Bind(ctx echo.Context, obj interface{}) error {
	if ctx.Request() != nil && ctx.Request().ContentLength > 0 {
		return json.NewDecoder(ctx.Request().Body).Decode(obj)
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
