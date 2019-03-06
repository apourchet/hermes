package hermes

import (
	"encoding/json"
	"fmt"
)

const (
	HERMES_CODE_BYPASS = -1
)

func findEndpointByHandler(svc Server, name string) (*Endpoint, error) {
	for _, ep := range svc.Endpoints() {
		if ep.Handler == name {
			return ep, nil
		}
	}
	return nil, fmt.Errorf("MethodNotFoundError")
}

func shouldJSON(obj interface{}) []byte {
	content, _ := json.Marshal(obj)
	return content
}
