package config

import (
	"encoding/json"
	"reflect"
	"testing"

	"k8s-firstcommit/pkg/api"
)

const TomcatContainerEtcdKey = "/registry/services/tomcat/endpoints/tomcat-3bd5af34"
const TomcatService = "tomcat"
const TomcatContainerId = "tomcat-3bd5af34"

func ValidateJsonParsing(t *testing.T, jsonString string, expectedEndpoints api.Endpoints, expectError bool) {
	endpoints, err := ParseEndpoints(jsonString)
	if err == nil && expectError {
		t.Errorf("ValidateJsonParsing did not get expected error when parsing %s", jsonString)
	}
	if err != nil && !expectError {
		t.Errorf("ValidateJsonParsing got unexpected error %+v when parsing %s", err, jsonString)
	}
	if !reflect.DeepEqual(expectedEndpoints, endpoints) {
		t.Errorf("Didn't get expected endpoints %+v got: %+v", expectedEndpoints, endpoints)
	}
}

func TestParseJsonEndpoints(t *testing.T) {
	ValidateJsonParsing(t, "", api.Endpoints{}, true)
	endpoints := api.Endpoints{
		Name:      "foo",
		Endpoints: []string{"foo", "bar", "baz"},
	}
	data, err := json.Marshal(endpoints)
	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
	ValidateJsonParsing(t, string(data), endpoints, false)
	//	ValidateJsonParsing(t, "[{\"port\":8000,\"name\":\"mysql\",\"machine\":\"foo\"},{\"port\":9000,\"name\":\"mysql\",\"machine\":\"bar\"}]", []string{"foo:8000", "bar:9000"}, false)
}
