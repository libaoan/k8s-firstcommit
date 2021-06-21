package util

import (
	"io/ioutil"
	"log"
	"net/http"
	"testing"
)

// FakeHanlder is to assist in testing HTTP requests
type FakeHandler struct {
	RequestReceived *http.Request
	StatusCode      int
	ResponseBody    string
}

func (f *FakeHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	f.RequestReceived = request
	response.WriteHeader(f.StatusCode)
	response.Write([]byte(f.ResponseBody))

	bodyReceived, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("Received read error: %#v", err)
	}
	f.ResponseBody = string(bodyReceived)
}

func (f *FakeHandler) ValidateRequest(t *testing.T, expectedPath, expectedMethod string, body *string) {
	if f.RequestReceived.URL.Path != expectedPath {
		t.Errorf("Unexpected request path: %s", f.RequestReceived.URL.Path)
	}
	if f.RequestReceived.Method != expectedMethod {
		t.Errorf("Unexpected request method: %s", f.RequestReceived.Method)
	}
	if body != nil {
		if *body != f.ResponseBody {
			t.Errorf("Received body:\n%s\n Doesn't match expected body:\n%s", f.ResponseBody, *body)
		}
	}
}
