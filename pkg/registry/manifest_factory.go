package registry

import (
	"k8s-firstcommit/pkg/api"
)

type ManifestFactory interface {
	// Make a container object for a given task, given the machine that the task is running on.
	MakeManifest(machine string, task api.Task) (api.ContainerManifest, error)
}

type BasicManifestFactory struct {
	serviceRegistry ServiceRegistry
}

func (b *BasicManifestFactory) MakeManifest(machine string, task api.Task) (api.ContainerManifest, error) {
	envVars, err := GetServiceEnvironmentVariables(b.serviceRegistry, machine)
	if err != nil {
		return api.ContainerManifest{}, err
	}
	for ix, container := range task.DesiredState.Manifest.Containers {
		task.DesiredState.Manifest.Id = task.ID
		task.DesiredState.Manifest.Containers[ix].Env = append(container.Env, envVars...)
	}
	return task.DesiredState.Manifest, nil
}
