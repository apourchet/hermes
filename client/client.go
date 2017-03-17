package client

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
	Exec(ctx context.Context, url, method string, in io.Reader, out interface{}) error
}

var _ IClient = &Client{}
var _ IClient = &MockClient{}

var DefaultClient IClient = &Client{"http", &http.Client{}}

type Client struct {
	Scheme     string
	HttpClient *http.Client
}

func (c *Client) Exec(ctx context.Context, url, method string, in io.Reader, out interface{}) error {
	req, err := http.NewRequest(method, fmt.Sprintf("%s://%s", c.Scheme, url), in)
	if err != nil {
		return err
	}

	resp, err := ctxhttp.Do(ctx, c.HttpClient, req)
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

func (c *MockClient) Exec(ctx context.Context, url, method string, in io.Reader, out interface{}) error {
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
