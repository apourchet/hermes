package binding

import (
	"net/http"
)

type SequentialBinding []Binding

func NewSequentialBinding(bindings ...Binding) SequentialBinding {
	return bindings
}

func (bindings SequentialBinding) Bind(req *http.Request, obj interface{}) error {
	for _, b := range bindings {
		if err := b.Bind(req, obj); err != nil {
			return err
		}
	}
	return nil
}

func (bindings SequentialBinding) Apply(obj interface{}, req *http.Request) error {
	for _, b := range bindings {
		if err := b.Apply(obj, req); err != nil {
			return err
		}
	}
	return nil
}
