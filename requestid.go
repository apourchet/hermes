package hermes

import (
	"net/http"

	uuid "github.com/satori/go.uuid"
)

const HermesRIDHeader = "Hermes-Request-ID"

func GetRequestID(req *http.Request) string {
	return req.Header.Get(HermesRIDHeader)
}

func SetRequestID(req *http.Request, rid string) *http.Request {
	req.Header.Set(HermesRIDHeader, rid)
	return req
}

func EnsureRequestID(req *http.Request) {
	rid := req.Header.Get(HermesRIDHeader)
	if rid == "" {
		rid = uuid.NewV4().String()
	}
	req.Header.Set(HermesRIDHeader, rid)
}

func TransferRequestID(src *http.Request, dst *http.Request) {
	rid := GetRequestID(src)
	dst.Header.Set("Hermes-Request-ID", rid)
}
