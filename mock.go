package hermes

import (
	"context"
	"fmt"
	"reflect"
)

// Any Serviceable is Mockable
type Mockable interface {
	Server
	// Should also implement all endpoints
}

type MockService struct {
	mockable Mockable
}

func NewMockService(mock Mockable) *MockService {
	return &MockService{mock}
}

func (m *MockService) Call(ctx context.Context, name string, in, out interface{}) error {
	svc := m.mockable
	_, err := findEndpointByHandler(svc, name)
	if err != nil {
		return err
	}

	serviceType := reflect.TypeOf(svc)
	method, ok := serviceType.MethodByName(name)
	if !ok {
		return fmt.Errorf("MethodNotImplementedError: %s", name)
	}

	args := []reflect.Value{reflect.ValueOf(svc), reflect.ValueOf(ctx), reflect.ValueOf(in), reflect.ValueOf(out)}
	vals := method.Func.Call(args)
	if !vals[1].IsNil() {
		errVal := vals[1].Interface().(error)
		return errVal
	}
	return nil
}
