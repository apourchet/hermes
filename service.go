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
	Client   IClient
	Resolve  Resolver
	Bindings BindingFactorySource

	serviceable Serviceable
}

func NewService(svc Serviceable) *Service {
	return &Service{DefaultClient, DefaultResolver, DefaultBindingFactorySource, svc}
}

func (s *Service) Call(ctx context.Context, name string, in, out interface{}) error {
	svc := s.serviceable
	ep, err := findEndpointByHandler(svc, name)
	if err != nil {
		return err
	}

	url, err := s.Resolve(svc.SNI(), ep.Path)
	if err != nil {
		return err
	}

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
		s.serveEndpoint(e, ep, method)
	}
	return nil
}

func (s *Service) serveEndpoint(e *gin.Engine, ep Endpoint, method reflect.Method) {
	binding := s.Bindings(ep.Path)
	fn := getGinHandler(s.serviceable, binding, ep, method)
	e.Handle(ep.Method, ep.Path, fn)
}
