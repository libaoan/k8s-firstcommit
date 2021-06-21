package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type ContainerInfo interface {
	GetContainerInfo(host, name string) (interface{}, error)
}

type HTTPContainerInfo struct {
	Client *http.Client
	Port   uint
}

func (c *HTTPContainerInfo) GetContainerInfo(host, name string) (interface{}, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("http://%s:%d/containerInfo?container=%s", host, c.Port, name), nil)
	if err != nil {
		return nil, err
	}
	response, err := c.Client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var data interface{}
	err = json.Unmarshal(body, &data)
	return data, err
}

// Useful for testing.
type FakeContainerInfo struct {
	data interface{}
	err  error
}

func (c *FakeContainerInfo) GetContainerInfo(host, name string) (interface{}, error) {
	return c.data, c.err
}
