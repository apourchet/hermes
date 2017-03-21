package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"

	"golang.org/x/net/context/ctxhttp"
)

type IClient interface {
	Exec(ctx context.Context, req *http.Request) (*http.Response, error)
}

var _ IClient = &Client{}
var _ IClient = &MockClient{}

var DefaultClient IClient = &Client{&http.Client{}}

type Client struct {
	HttpClient *http.Client
}

func (c *Client) Exec(ctx context.Context, req *http.Request) (*http.Response, error) {
	return ctxhttp.Do(ctx, c.HttpClient, req)
}

type MockClient struct {
	*gin.Engine
}

func (c *MockClient) Exec(ctx context.Context, req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()
	c.ServeHTTP(w, req)
	resp := w.Result()
	if resp == nil {
		return nil, fmt.Errorf("Failed to get http response")
	}
	return resp, nil
}
