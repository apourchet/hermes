package hermes

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/apourchet/hermes/client"
	"github.com/apourchet/hermes/resolver"
	"github.com/gin-gonic/gin"
)

type Service struct {
	Client   client.IClient
	Resolve  resolver.Resolver
	Bindings BindingFactory

	Scheme string

	serviceable IServiceable
}

func NewService(svc IServiceable) *Service {
	out := &Service{}
	out.Client = client.DefaultClient
	out.Resolve = resolver.DefaultResolver
	out.Bindings = DefaultBindingFactory

	out.Scheme = "http"

	out.serviceable = svc
	return out
}

func (s *Service) Call(ctx context.Context, name string, in, out interface{}) error {
	svc := s.serviceable
	// Get endpoint
	ep, err := findEndpointByHandler(svc, name)
	if err != nil {
		return fmt.Errorf("Client failed to find endpoint: %v", err)
	}

	// Resolve URL
	url, err := s.Resolve(svc.SNI(), ep.Path)
	if err != nil {
		return fmt.Errorf("Client failed to resolve url: %v", err)
	}

	// Create new request
	req, err := http.NewRequest(ep.Method, fmt.Sprintf("%s://%s", s.Scheme, url), nil)
	if err != nil {
		return fmt.Errorf("Client failed to create new http request")
	}

	// Use bindings on request
	err = s.Bindings(ep.Params, ep.Queries, ep.Headers).Apply(req, in)
	if err != nil {
		return fmt.Errorf("Client failed to apply a binding: %v", err)
	}

	// Execute request
	resp, err := s.Client.Exec(ctx, req)
	if err != nil {
		return fmt.Errorf("Client failed execute request: %v", err)
	}
	defer resp.Body.Close()

	// Read in response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Client failed read response body: %v", err)
	}

	// Deal with response
	if resp.StatusCode/100 == 2 {
		if out != nil {
			if err := json.Unmarshal(body, out); err != nil {
				return fmt.Errorf("Client failed to unmarshal response into output: %v", err)
			}
		}
		return nil
	}

	// There was an error
	tmp := map[string]string{}
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		return fmt.Errorf("Client failed to parse error response: %v", err)
	}
	if message, found := tmp["message"]; found {
		return fmt.Errorf(message)
	}
	return fmt.Errorf("Client failed to find error message. Status code was %d.", resp.StatusCode)
}

func (svc *Service) Serve(engine *gin.Engine) error {
	return NewServable(svc.serviceable, svc.Bindings).Serve(engine)
}
