package hermes

import (
	"encoding/json"
	"net/http"

	"github.com/golang/glog"
)

type Error struct {
	Message string
}

func (e Error) Error() string {
	return e.Message
}

func (e Error) JSON() []byte {
	content, _ := json.Marshal(e)
	return content
}

type ErrorHandler func(req *http.Request, code int, err error)

type SuccessHandler func(req *http.Request, code int)

var DefaultErrorHandler ErrorHandler = LogError

func LogError(req *http.Request, code int, err error) {
	glog.Errorf("[%s] %s => %d | %v", GetRequestID(req), req.URL.Path, code, err)
}

var DefaultSuccessHandler SuccessHandler = func(req *http.Request, code int) {
	glog.Infof("[%s] %s => %d", GetRequestID(req), req.URL.Path, code)
}
