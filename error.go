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

type ErrorHandler func(context.Context, error)

var DefaultErrorHandler ErrorHandler = LogError

func LogError(ctx context.Context, err error) {
	glog.Errorf("%v", err)
}
