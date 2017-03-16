package hermes

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"

	"golang.org/x/net/context/ctxhttp"
)

type IClient interface {
	GetScheme() string
	SetScheme(string) string
	Do(ctx context.Context, url, method string, in io.Reader, out interface{}) error
}

var _ IClient = &Client{}
var _ IClient = &MockClient{}

var DefaultClient IClient = &Client{"http"}

type Client struct {
	Scheme string
}

func (c *Client) GetScheme() string {
	return c.Scheme
}

func (c *Client) SetScheme(scheme string) string {
	old := c.Scheme
	c.Scheme = scheme
	return old
}

func (c *Client) Do(ctx context.Context, url, method string, in io.Reader, out interface{}) error {
	req, err := http.NewRequest(method, fmt.Sprintf("%s://%s", c.Scheme, url), in)
	if err != nil {
		return err
	}

	cli := &http.Client{}
	resp, err := ctxhttp.Do(ctx, cli, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Cound not read body")
		return err
	}

	if resp.StatusCode/100 == 2 {
		if out != nil {
			err = json.Unmarshal(body, out)
			return err
		}
		return nil
	}

	tmp := map[string]string{}
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		return err
	}
	if message, found := tmp["message"]; found {
		return fmt.Errorf(message)
	}
	return fmt.Errorf("Malformatted error in handler. Status code was %d.", resp.StatusCode)
}

type MockClient struct {
	Scheme string

	*gin.Engine
}

func (c *MockClient) GetScheme() string {
	return c.Scheme
}

func (c *MockClient) SetScheme(scheme string) string {
	old := c.Scheme
	c.Scheme = scheme
	return old
}

func (c *MockClient) Do(ctx context.Context, url, method string, in io.Reader, out interface{}) error {
	req, err := http.NewRequest(method, fmt.Sprintf("%s://%s", c.Scheme, url), in)
	if err != nil {
		return err
	}

	w := httptest.NewRecorder()
	c.ServeHTTP(w, req)
	resp := w.Result()
	if resp == nil {
		return fmt.Errorf("Failed to get http response")
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode/100 == 2 {
		if out != nil {
			err = json.Unmarshal(body, out)
			return err
		}
		return nil
	}

	tmp := map[string]string{}
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		return err
	}
	if message, found := tmp["message"]; found {
		return fmt.Errorf(message)
	}
	return fmt.Errorf("Malformatted error in handler. Status code was %d.", resp.StatusCode)
}
