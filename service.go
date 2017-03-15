package hermes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/gin-gonic/gin"
)

type Service struct {
	serviceable Serviceable
}

func NewService(svc Serviceable) *Service {
	return &Service{svc}
}

func (s *Service) Call(ctx context.Context, name string, in, out interface{}) error {
	svc := s.serviceable
	ep, err := findEndpointByHandler(svc, name)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("http://%s%s", svc.Host(), ep.Path)

	inData, err := json.Marshal(in)
	if err != nil {
		return err
	}

	return DefaultClient.Do(ctx, url, ep.Method, bytes.NewBuffer(inData), out)
}

func (s *Service) Serve(e *gin.Engine) error {
	svc := s.serviceable
	serviceType := reflect.TypeOf(svc)
	endpoints := svc.Endpoints()
	for _, ep := range endpoints {
		method, ok := serviceType.MethodByName(ep.Handler)
		if !ok {
			return fmt.Errorf("Endpoint '%s' does not match any method of the type %t", ep.Handler, serviceType)
		}
		serveEndpoint(e, svc, ep, method)
	}
	return nil
}
