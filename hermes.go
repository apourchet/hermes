package hermes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
)

// Input types. These interfaces must be satisfied by the user
type Serviceable interface {
	Hosted
	Server
	// Should also implement all endpoints
}

// Any Serviceable is Mockable
type Mockable interface {
	Server
	// Should also implement all endpoints
}

type EndpointMap []Endpoint

type Endpoint struct {
	Handler   string
	Method    string
	Path      string
	NewInput  func() interface{}
	NewOutput func() interface{}
}

type Hosted interface {
	Host() string
}

type Server interface {
	Endpoints() EndpointMap
}

// Output types. These interfaces are satisfied by
// our outputs
type ServiceInterface interface {
	Serve(e *gin.Engine) error
	Callable
}

type Callable interface {
	Call(name string, in, out interface{}) (int, error)
}

type service struct {
	serviceable Serviceable
}

type mockService struct {
	mockable Mockable
}

// Implementations

func (s *service) Call(name string, in, out interface{}) (int, error) {
	svc := s.serviceable
	ep, err := findEndpointByHandler(svc, name)
	if err != nil {
		return 404, err
	}
	url := fmt.Sprintf("http://%s%s", svc.Host(), ep.Path)
	// log.Printf("RPC Call (%s) => %s(%s)", name, ep.Method, url)

	inData, err := json.Marshal(in)
	if err != nil {
		return 400, err
	}

	req, err := http.NewRequest(ep.Method, url, bytes.NewBuffer(inData))
	if err != nil {
		return 400, err
	}

	resp, err := getDefaultClient().Do(req)
	if err != nil {
		return 400, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Cound not read body")
		return resp.StatusCode, err
	}

	err = json.Unmarshal(body, out)
	if err != nil {
		return resp.StatusCode, err
	}
	return resp.StatusCode, nil
}

func getDefaultClient() *http.Client {
	return &http.Client{}
}

func findEndpointByHandler(svc Server, name string) (Endpoint, error) {
	for _, ep := range svc.Endpoints() {
		if ep.Handler == name {
			return ep, nil
		}
	}
	return Endpoint{}, fmt.Errorf("MethodNotFoundError")
}

func (m *mockService) Call(name string, in, out interface{}) (int, error) {
	svc := m.mockable
	_, err := findEndpointByHandler(svc, name)
	if err != nil {
		return 404, err
	}

	serviceType := reflect.TypeOf(svc)
	method, ok := serviceType.MethodByName(name)
	if !ok {
		return 404, fmt.Errorf("MethodNotImplementedError: %s", name)
	}

	args := []reflect.Value{reflect.ValueOf(svc), reflect.ValueOf(in), reflect.ValueOf(out)}
	vals := method.Func.Call(args)
	code := int(vals[0].Int())
	if !vals[1].IsNil() {
		errVal := vals[1].Interface().(error)
		return code, errVal
	}
	return code, nil
}

func (s *service) Serve(e *gin.Engine) error {
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

func serveEndpoint(e *gin.Engine, svc Serviceable, ep Endpoint, method reflect.Method) {
	fn := getGinHandler(svc, ep, method)
	e.Handle(ep.Method, ep.Path, fn)
}

func getGinHandler(svc Serviceable, ep Endpoint, method reflect.Method) func(c *gin.Context) {
	return func(c *gin.Context) {
		input := ep.NewInput()
		output := ep.NewOutput()
		err := c.BindJSON(input)
		if err != nil {
			c.JSON(http.StatusBadRequest, &gin.H{"error": err.Error()})
			log.Println(err)
			return
		}
		args := []reflect.Value{reflect.ValueOf(svc), reflect.ValueOf(c), reflect.ValueOf(input), reflect.ValueOf(output)}
		vals := method.Func.Call(args)
		code := int(vals[0].Int())
		if !vals[1].IsNil() {
			errVal := vals[1].Interface().(error)
			c.JSON(code, gin.H{"error": errVal.Error()})
		} else {
			c.JSON(code, output)
		}
	}
}

func InitService(svc Serviceable) ServiceInterface {
	return &service{svc}
}

func InitMockService(mock Mockable) Callable {
	return &mockService{mock}
}
