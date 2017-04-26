package hermes

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

func GetRequestID(ctx context.Context) string {
	rid := ctx.Value("Hermes-Request-ID")
	if rid == nil {
		return ""
	}
	return rid.(string)
}

func SetRequestID(ctx context.Context, rid string) context.Context {
	return context.WithValue(ctx, "Hermes-Request-ID", rid)
}

func EnsureRequestID(ctx *gin.Context) {
	rid := ctx.Request.Header.Get("Hermes-Request-ID")
	if rid == "" {
		rid = uuid.NewV4().String()
	}
	ctx.Set("Hermes-Request-ID", rid)
}

func TransferRequestID(ctx context.Context, req *http.Request) {
	rid := GetRequestID(ctx)
	req.Header.Set("Hermes-Request-ID", rid)
}
