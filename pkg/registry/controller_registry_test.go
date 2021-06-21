package registry

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"k8s-firstcommit/pkg"
	"reflect"
	"testing"

	. "k8s-firstcommit/pkg/api"
)

type MockControllerRegistry struct {
	err         error
	controllers []ReplicationController
}

func (registry *MockControllerRegistry) ListControllers() ([]ReplicationController, error) {
	return registry.controllers, registry.err
}

func (registry *MockControllerRegistry) GetController(ID string) (*ReplicationController, error) {
	return &ReplicationController{}, registry.err
}

func (registry *MockControllerRegistry) CreateController(controller ReplicationController) error {
	return registry.err
}

func (registry *MockControllerRegistry) UpdateController(controller ReplicationController) error {
	return registry.err
}
func (registry *MockControllerRegistry) DeleteController(ID string) error {
	return registry.err
}

func TestListControllersError(t *testing.T) {
	mockRegistry := MockControllerRegistry{
		err: fmt.Errorf("Test Error"),
	}
	storage := pkg.ControllerRegistryStorage{
		registry: &mockRegistry,
	}
	controllersObj, err := storage.List(nil)
	controllers := controllersObj.(ReplicationControllerList)
	if err != mockRegistry.err {
		t.Errorf("Expected %#v, Got %#v", mockRegistry.err, err)
	}
	if len(controllers.Items) != 0 {
		t.Errorf("Unexpected non-zero task list: %#v", controllers)
	}
}

func TestListEmptyControllerList(t *testing.T) {
	mockRegistry := MockControllerRegistry{}
	storage := pkg.ControllerRegistryStorage{
		registry: &mockRegistry,
	}
	controllers, err := storage.List(nil)
	pkg.expectNoError(t, err)
	if len(controllers.(ReplicationControllerList).Items) != 0 {
		t.Errorf("Unexpected non-zero task list: %#v", controllers)
	}
}

func TestListControllerList(t *testing.T) {
	mockRegistry := MockControllerRegistry{
		controllers: []ReplicationController{
			ReplicationController{
				JSONBase: JSONBase{
					ID: "foo",
				},
			},
			ReplicationController{
				JSONBase: JSONBase{
					ID: "bar",
				},
			},
		},
	}
	storage := pkg.ControllerRegistryStorage{
		registry: &mockRegistry,
	}
	controllersObj, err := storage.List(nil)
	controllers := controllersObj.(ReplicationControllerList)
	pkg.expectNoError(t, err)
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
	storage := pkg.ControllerRegistryStorage{
		registry: &mockRegistry,
	}
	controller := ReplicationController{
		JSONBase: JSONBase{
			ID: "foo",
		},
	}
	body, err := json.Marshal(controller)
	pkg.expectNoError(t, err)
	controllerOut, err := storage.Extract(string(body))
	pkg.expectNoError(t, err)
	jsonOut, err := json.Marshal(controllerOut)
	pkg.expectNoError(t, err)
	if string(body) != string(jsonOut) {
		t.Errorf("Expected %#v, found %#v", controller, controllerOut)
	}
}

func TestControllerParsing(t *testing.T) {
	expectedController := ReplicationController{
		JSONBase: JSONBase{
			ID: "nginxController",
		},
		DesiredState: ReplicationControllerState{
			Replicas: 2,
			ReplicasInSet: map[string]string{
				"name": "nginx",
			},
			TaskTemplate: TaskTemplate{
				DesiredState: TaskState{
					Manifest: ContainerManifest{
						Containers: []Container{
							Container{
								Image: "dockerfile/nginx",
								Ports: []Port{
									Port{
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
	pkg.expectNoError(t, err)
	data, err := json.Marshal(expectedController)
	pkg.expectNoError(t, err)
	_, err = file.Write(data)
	pkg.expectNoError(t, err)
	err = file.Close()
	pkg.expectNoError(t, err)
	data, err = ioutil.ReadFile(fileName)
	pkg.expectNoError(t, err)
	var controller ReplicationController
	err = json.Unmarshal(data, &controller)
	pkg.expectNoError(t, err)

	if !reflect.DeepEqual(controller, expectedController) {
		t.Errorf("Parsing failed: %s %#v %#v", string(data), controller, expectedController)
	}
}
