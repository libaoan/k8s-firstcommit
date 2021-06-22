package registry

import (
	"testing"

	"k8s-firstcommit/pkg/api"
)

func TestMakeManifestNoServices(t *testing.T) {
	registry := MockServiceRegistry{}
	factory := &BasicManifestFactory{
		serviceRegistry: &registry,
	}

	manifest, err := factory.MakeManifest("machine", api.Task{
		JSONBase: api.JSONBase{ID: "foobar"},
		DesiredState: api.TaskState{
			Manifest: api.ContainerManifest{
				Containers: []api.Container{
					api.Container{
						Name: "foo",
					},
				},
			},
		},
	})
	expectNoError(t, err)
	container := manifest.Containers[0]
	if len(container.Env) != 1 ||
		container.Env[0].Name != "SERVICE_HOST" ||
		container.Env[0].Value != "machine" {
		t.Errorf("Expected one env vars, got: %#v", manifest)
	}
	if manifest.Id != "foobar" {
		t.Errorf("Failed to assign id to manifest: %#v")
	}
}

func TestMakeManifestServices(t *testing.T) {
	registry := MockServiceRegistry{
		list: api.ServiceList{
			Items: []api.Service{
				api.Service{
					JSONBase: api.JSONBase{ID: "test"},
					Port:     8080,
				},
			},
		},
	}
	factory := &BasicManifestFactory{
		serviceRegistry: &registry,
	}

	manifest, err := factory.MakeManifest("machine", api.Task{
		DesiredState: api.TaskState{
			Manifest: api.ContainerManifest{
				Containers: []api.Container{
					api.Container{
						Name: "foo",
					},
				},
			},
		},
	})
	expectNoError(t, err)
	container := manifest.Containers[0]
	if len(container.Env) != 2 ||
		container.Env[0].Name != "TEST_SERVICE_PORT" ||
		container.Env[0].Value != "8080" ||
		container.Env[1].Name != "SERVICE_HOST" ||
		container.Env[1].Value != "machine" {
		t.Errorf("Expected 2 env vars, got: %#v", manifest)
	}
}

func TestMakeManifestServicesExistingEnvVar(t *testing.T) {
	registry := MockServiceRegistry{
		list: api.ServiceList{
			Items: []api.Service{
				api.Service{
					JSONBase: api.JSONBase{ID: "test"},
					Port:     8080,
				},
			},
		},
	}
	factory := &BasicManifestFactory{
		serviceRegistry: &registry,
	}

	manifest, err := factory.MakeManifest("machine", api.Task{
		DesiredState: api.TaskState{
			Manifest: api.ContainerManifest{
				Containers: []api.Container{
					api.Container{
						Env: []api.EnvVar{
							api.EnvVar{
								Name:  "foo",
								Value: "bar",
							},
						},
					},
				},
			},
		},
	})
	expectNoError(t, err)
	container := manifest.Containers[0]
	if len(container.Env) != 3 ||
		container.Env[0].Name != "foo" ||
		container.Env[0].Value != "bar" ||
		container.Env[1].Name != "TEST_SERVICE_PORT" ||
		container.Env[1].Value != "8080" ||
		container.Env[2].Name != "SERVICE_HOST" ||
		container.Env[2].Value != "machine" {
		t.Errorf("Expected no env vars, got: %#v", manifest)
	}
}
