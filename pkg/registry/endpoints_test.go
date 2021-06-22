package registry

import (
	"fmt"
	"testing"

	"k8s-firstcommit/pkg/api"
)

func TestSyncEndpointsEmpty(t *testing.T) {
	serviceRegistry := MockServiceRegistry{}
	taskRegistry := MockTaskRegistry{}

	endpoints := MakeEndpointController(&serviceRegistry, &taskRegistry)
	err := endpoints.SyncServiceEndpoints()
	expectNoError(t, err)
}

func TestSyncEndpointsError(t *testing.T) {
	serviceRegistry := MockServiceRegistry{
		err: fmt.Errorf("Test Error"),
	}
	taskRegistry := MockTaskRegistry{}

	endpoints := MakeEndpointController(&serviceRegistry, &taskRegistry)
	err := endpoints.SyncServiceEndpoints()
	if err != serviceRegistry.err {
		t.Errorf("Errors don't match: %#v %#v", err, serviceRegistry.err)
	}
}

func TestSyncEndpointsItems(t *testing.T) {
	serviceRegistry := MockServiceRegistry{
		list: api.ServiceList{
			Items: []api.Service{
				api.Service{
					Labels: map[string]string{
						"foo": "bar",
					},
				},
			},
		},
	}
	taskRegistry := MockTaskRegistry{
		tasks: []api.Task{
			api.Task{
				DesiredState: api.TaskState{
					Manifest: api.ContainerManifest{
						Containers: []api.Container{
							api.Container{
								Ports: []api.Port{
									api.Port{
										HostPort: 8080,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	endpoints := MakeEndpointController(&serviceRegistry, &taskRegistry)
	err := endpoints.SyncServiceEndpoints()
	expectNoError(t, err)
	if len(serviceRegistry.endpoints.Endpoints) != 1 {
		t.Errorf("Unexpected endpoints update: %#v", serviceRegistry.endpoints)
	}
}

func TestSyncEndpointsTaskError(t *testing.T) {
	serviceRegistry := MockServiceRegistry{
		list: api.ServiceList{
			Items: []api.Service{
				api.Service{
					Labels: map[string]string{
						"foo": "bar",
					},
				},
			},
		},
	}
	taskRegistry := MockTaskRegistry{
		err: fmt.Errorf("test error."),
	}

	endpoints := MakeEndpointController(&serviceRegistry, &taskRegistry)
	err := endpoints.SyncServiceEndpoints()
	if err == nil {
		t.Error("Unexpected non-error")
	}
}
