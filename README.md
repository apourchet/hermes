# Hermes
Hermes is a simple pseudo-RPC framework in golang. It's built on top of [gin](https://github.com/gin-gonic/gin). The main advantage is that it's 1) curlable and 2) has no code generation.

# Example
### Service Definition
```go
// Service definition
type MyService struct{}

func (s *MyService) SNI() string {
	return "localhost:9000"
}

func (s *MyService) Endpoints() hermes.EndpointMap {
	return hermes.EndpointMap{
		hermes.EP("RpcCall", "GET", "/test", Inbount{}, Outbound{}), 
		hermes.EP("OtherRpcCall", "POST", "/test", OtherInbound{}, OtherOutbound{}),
	}
}

// Endpoint definitions
// RpcCall
type Inbound struct { Message string }
type Outbound struct { Ok bool }

func (s *MyService) RpcCall(c *gin.Context, in *Inbound, out *Outbound) (int, error) {
	if in.Message == "secret" {
		out.Ok = true
		return http.StatusOK, nil
	}
	out.Ok = false
	return http.StatusBadRequest, fmt.Errorf("Wrong secret!")
}

// OtherRpcCall
type OtherInbound struct { MyField int }
type OtherOutbound struct { SomeFloat float64 }

func (s *MyService) OtherRpcCall(c *gin.Context, in *OtherInbound, out *OtherOutbound) (int, error) {
	out.SomeFloat = 3.14 * in.MyField
  	return http.StatusOK, nil
}
```

### Server Creation
```go
engine := gin.New()
server := hermes.NewServer(&MyService{})
server.Serve(engine)
engine.Run(":9000")
```

### Client RPC Call
```go
caller := hermes.NewCaller(&MyService{})
out := &Outbound{false}
code, err := caller.Call("RpcCall", &Inbound{"secret"}, out)
```
