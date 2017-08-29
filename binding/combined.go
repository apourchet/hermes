package binding

import (
	"net/http"

	"github.com/labstack/echo"
)

type SequentialBinding []Binding

func NewSequentialBinding(bindings ...Binding) SequentialBinding {
	return bindings
}

func (bindings SequentialBinding) Bind(ctx echo.Context, obj interface{}) error {
	for _, b := range bindings {
		if err := b.Bind(ctx, obj); err != nil {
			return err
		}
	}
	return nil
}

func (bindings SequentialBinding) Apply(req *http.Request, obj interface{}) error {
	for _, b := range bindings {
		if err := b.Apply(req, obj); err != nil {
			return err
		}
	}
	return nil
}
