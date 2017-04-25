package hermes

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type ICaller interface {
	Call(ctx context.Context, methodname string, in, out interface{}) (int, error)
}

type Caller struct {
	Client   IClient
	Resolve  Resolver
	Bindings BindingFactory

	Scheme string

	callable ICallable
}

func NewCaller(callable ICallable) *Caller {
	out := &Caller{}
	out.Client = DefaultClient
	out.Resolve = DefaultResolver
	out.Bindings = DefaultBindingFactory
	out.Scheme = "http"
	out.callable = callable
	return out
}

func (caller *Caller) Call(ctx context.Context, methodname string, in, out interface{}) (int, error) {
	callable := caller.callable

	// Get endpoint
	ep, err := findEndpointByHandler(callable, methodname)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("Client failed to find endpoint: %v", err)
	}

	// Resolve URL
	url, err := caller.Resolve(callable.SNI(), ep.Path)
	if err != nil {
		return http.StatusNotFound, fmt.Errorf("Client failed to resolve url: %v", err)
	}

	// Create new request
	req, err := http.NewRequest(ep.Method, fmt.Sprintf("%s://%s", caller.Scheme, url), nil)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("Client failed to create new http request")
	}

	// Use bindings on request
	err = caller.Bindings(ep.Params, ep.Queries, ep.Headers).Apply(req, in)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("Client failed to apply a binding: %v", err)
	}

	// Transfer request ID to call
	TransferRequestID(ctx, req)

	// Execute request
	resp, err := caller.Client.Exec(ctx, req)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("Client failed execute request: %v", err)
	}
	defer resp.Body.Close()

	// Read in response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("Client failed read response body: %v", err)
	}

	// Deal with response
	if resp.StatusCode/100 == 2 {
		if out != nil {
			if err := json.Unmarshal(body, out); err != nil {
				return resp.StatusCode, fmt.Errorf("Client failed to unmarshal response into output: %v", err)
			}
		}
		return resp.StatusCode, nil
	}

	// There was an error
	tmp := &Error{}
	err = json.Unmarshal(body, tmp)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("Client failed to parse error response: %v", err)
	}
	return resp.StatusCode, tmp
}
