package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"k8s-firstcommit/pkg/util"
)

func TestHTTPContainerInfo(t *testing.T) {
	body := `{"items":[]}`
	fakeHandler := util.FakeHandler{
		StatusCode:   200,
		ResponseBody: body,
	}
	testServer := httptest.NewServer(&fakeHandler)

	hostUrl, err := url.Parse(testServer.URL)
	expectNoError(t, err)
	parts := strings.Split(hostUrl.Host, ":")

	port, err := strconv.Atoi(parts[1])
	expectNoError(t, err)
	containerInfo := &HTTPContainerInfo{
		Client: http.DefaultClient,
		Port:   uint(port),
	}
	data, err := containerInfo.GetContainerInfo(parts[0], "foo")
	expectNoError(t, err)
	dataString, _ := json.Marshal(data)
	if string(dataString) != body {
		t.Errorf("Unexpected response.  Expected: %s, received %s", body, string(dataString))
	}
}
