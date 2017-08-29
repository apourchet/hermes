package binding

import (
	"net/http"

	"github.com/labstack/echo"
)

type Binding interface {
	Bind(ctx echo.Context, obj interface{}) error
	Apply(req *http.Request, obj interface{}) error
}
