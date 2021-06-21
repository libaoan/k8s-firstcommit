package registry

import (
	"fmt"
	"k8s-firstcommit/pkg"
	"testing"

	. "k8s-firstcommit/pkg/api"
)

func TestSyncEndpointsEmpty(t *testing.T) {
	serviceRegistry := pkg.MockServiceRegistry{}
	taskRegistry := pkg.MockTaskRegistry{}

	endpoints := pkg.MakeEndpointController(&serviceRegistry, &taskRegistry)
	err := endpoints.SyncServiceEndpoints()
	pkg.expectNoError(t, err)
}

func TestSyncEndpointsError(t *testing.T) {
	serviceRegistry := pkg.MockServiceRegistry{
		err: fmt.Errorf("Test Error"),
	}
	taskRegistry := pkg.MockTaskRegistry{}

	endpoints := pkg.MakeEndpointController(&serviceRegistry, &taskRegistry)
	err := endpoints.SyncServiceEndpoints()
	if err != serviceRegistry.err {
		t.Errorf("Errors don't match: %#v %#v", err, serviceRegistry.err)
	}
}

func TestSyncEndpointsItems(t *testing.T) {
	serviceRegistry := pkg.MockServiceRegistry{
		list: ServiceList{
			Items: []Service{
				Service{
					Labels: map[string]string{
						"foo": "bar",
					},
				},
			},
		},
	}
	taskRegistry := pkg.MockTaskRegistry{
		tasks: []Task{
			Task{
				DesiredState: TaskState{
					Manifest: ContainerManifest{
						Containers: []Container{
							Container{
								Ports: []Port{
									Port{
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

	endpoints := pkg.MakeEndpointController(&serviceRegistry, &taskRegistry)
	err := endpoints.SyncServiceEndpoints()
	pkg.expectNoError(t, err)
	if len(serviceRegistry.endpoints.Endpoints) != 1 {
		t.Errorf("Unexpected endpoints update: %#v", serviceRegistry.endpoints)
	}
}

func TestSyncEndpointsTaskError(t *testing.T) {
	serviceRegistry := pkg.MockServiceRegistry{
		list: ServiceList{
			Items: []Service{
				Service{
					Labels: map[string]string{
						"foo": "bar",
					},
				},
			},
		},
	}
	taskRegistry := pkg.MockTaskRegistry{
		err: fmt.Errorf("test error."),
	}

	endpoints := pkg.MakeEndpointController(&serviceRegistry, &taskRegistry)
	err := endpoints.SyncServiceEndpoints()
	if err == nil {
		t.Error("Unexpected non-error")
	}
}
