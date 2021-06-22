package registry

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"

	"k8s-firstcommit/pkg/api"
)

type MockControllerRegistry struct {
	err         error
	controllers []api.ReplicationController
}

func (registry *MockControllerRegistry) ListControllers() ([]api.ReplicationController, error) {
	return registry.controllers, registry.err
}

func (registry *MockControllerRegistry) GetController(ID string) (*api.ReplicationController, error) {
	return &api.ReplicationController{}, registry.err
}

func (registry *MockControllerRegistry) CreateController(controller api.ReplicationController) error {
	return registry.err
}

func (registry *MockControllerRegistry) UpdateController(controller api.ReplicationController) error {
	return registry.err
}
func (registry *MockControllerRegistry) DeleteController(ID string) error {
	return registry.err
}

func TestListControllersError(t *testing.T) {
	mockRegistry := MockControllerRegistry{
		err: fmt.Errorf("Test Error"),
	}
	storage := ControllerRegistryStorage{
		registry: &mockRegistry,
	}
	controllersObj, err := storage.List(nil)
	controllers := controllersObj.(api.ReplicationControllerList)
	if err != mockRegistry.err {
		t.Errorf("Expected %#v, Got %#v", mockRegistry.err, err)
	}
	if len(controllers.Items) != 0 {
		t.Errorf("Unexpected non-zero task list: %#v", controllers)
	}
}

func TestListEmptyControllerList(t *testing.T) {
	mockRegistry := MockControllerRegistry{}
	storage := ControllerRegistryStorage{
		registry: &mockRegistry,
	}
	controllers, err := storage.List(nil)
	expectNoError(t, err)
	if len(controllers.(api.ReplicationControllerList).Items) != 0 {
		t.Errorf("Unexpected non-zero task list: %#v", controllers)
	}
}

func TestListControllerList(t *testing.T) {
	mockRegistry := MockControllerRegistry{
		controllers: []api.ReplicationController{
			api.ReplicationController{
				JSONBase: api.JSONBase{
					ID: "foo",
				},
			},
			api.ReplicationController{
				JSONBase: api.JSONBase{
					ID: "bar",
				},
			},
		},
	}
	storage := ControllerRegistryStorage{
		registry: &mockRegistry,
	}
	controllersObj, err := storage.List(nil)
	controllers := controllersObj.(api.ReplicationControllerList)
	expectNoError(t, err)
	if len(controllers.Items) != 2 {
		t.Errorf("Unexpected controller list: %#v", controllers)
	}
	if controllers.Items[0].ID != "foo" {
		t.Errorf("Unexpected controller: %#v", controllers.Items[0])
	}
	if controllers.Items[1].ID != "bar" {
		t.Errorf("Unexpected controller: %#v", controllers.Items[1])
	}
}

func TestExtractControllerJson(t *testing.T) {
	mockRegistry := MockControllerRegistry{}
	storage := ControllerRegistryStorage{
		registry: &mockRegistry,
	}
	controller := api.ReplicationController{
		JSONBase: api.JSONBase{
			ID: "foo",
		},
	}
	body, err := json.Marshal(controller)
	expectNoError(t, err)
	controllerOut, err := storage.Extract(string(body))
	expectNoError(t, err)
	jsonOut, err := json.Marshal(controllerOut)
	expectNoError(t, err)
	if string(body) != string(jsonOut) {
		t.Errorf("Expected %#v, found %#v", controller, controllerOut)
	}
}

func TestControllerParsing(t *testing.T) {
	expectedController := api.ReplicationController{
		JSONBase: api.JSONBase{
			ID: "nginxController",
		},
		DesiredState: api.ReplicationControllerState{
			Replicas: 2,
			ReplicasInSet: map[string]string{
				"name": "nginx",
			},
			TaskTemplate: api.TaskTemplate{
				DesiredState: api.TaskState{
					Manifest: api.ContainerManifest{
						Containers: []api.Container{
							api.Container{
								Image: "dockerfile/nginx",
								Ports: []api.Port{
									api.Port{
										ContainerPort: 80,
										HostPort:      8080,
									},
								},
							},
						},
					},
				},
				Labels: map[string]string{
					"name": "nginx",
				},
			},
		},
		Labels: map[string]string{
			"name": "nginx",
		},
	}
	file, err := ioutil.TempFile("", "controller")
	fileName := file.Name()
	expectNoError(t, err)
	data, err := json.Marshal(expectedController)
	expectNoError(t, err)
	_, err = file.Write(data)
	expectNoError(t, err)
	err = file.Close()
	expectNoError(t, err)
	data, err = ioutil.ReadFile(fileName)
	expectNoError(t, err)
	var controller api.ReplicationController
	err = json.Unmarshal(data, &controller)
	expectNoError(t, err)

	if !reflect.DeepEqual(controller, expectedController) {
		t.Errorf("Parsing failed: %s %#v %#v", string(data), controller, expectedController)
	}
}
