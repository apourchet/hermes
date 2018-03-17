package hermes

import (
	"context"

	"github.com/golang/glog"
)

type Error struct {
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

type ErrorHandler func(ctx context.Context, path string, code int, err error)

type SuccessHandler func(ctx context.Context, path string, code int)

var DefaultErrorHandler ErrorHandler = LogError

func LogError(ctx context.Context, path string, code int, err error) {
	glog.Errorf("[%s] %s => %d | %v", GetRequestID(ctx), path, code, err)
}

var DefaultSuccessHandler SuccessHandler = func(ctx context.Context, path string, code int) {
	glog.Infof("[%s] %s => %d", GetRequestID(ctx), path, code)
}
