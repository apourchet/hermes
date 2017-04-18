package hermes

import (
	"context"

	"github.com/apourchet/hermes/requestid"
	"github.com/golang/glog"
)

type Error struct {
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

type ErrorHandler func(ctx context.Context, path string, code int, err error)

var DefaultErrorHandler ErrorHandler = LogError

func LogError(ctx context.Context, path string, code int, err error) {
	glog.Errorf("[%s] %s => %d | %v", requestid.GetRequestID(ctx), path, code, err)
}
