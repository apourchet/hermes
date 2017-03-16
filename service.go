package hermes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"github.com/gin-gonic/gin"
)

type Service struct {
	Client        IClient
	Resolve       Resolver
	Bindings      BindingFactorySource
	QueryTemplate QueryTemplater

	serviceable Serviceable
}

func NewService(svc Serviceable) *Service {
	out := &Service{}
	out.Client = DefaultClient
	out.Resolve = DefaultResolver
	out.Bindings = DefaultBindingFactorySource
	out.QueryTemplate = DefaultQueryTemplate

	out.serviceable = svc
	return out
}

func (s *Service) Call(ctx context.Context, name string, in, out interface{}) error {
	svc := s.serviceable
	ep, err := findEndpointByHandler(svc, name)
	if err != nil {
		return err
	}

	var body io.Reader
	if in != nil {
		inData, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewBuffer(inData)
	}

	fullpath, err := s.QueryTemplate(ep.Path, in)
	if err != nil {
		return err
	}

	url, err := s.Resolve(svc.SNI(), fullpath)
	if err != nil {
		return err
	}

	return s.Client.Do(ctx, url, ep.Method, body, out)
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