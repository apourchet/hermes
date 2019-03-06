package binding

import (
	"net/http"
)

type Binding interface {
	Bind(req *http.Request, obj interface{}) error
	Apply(obj interface{}, req *http.Request) error
}
