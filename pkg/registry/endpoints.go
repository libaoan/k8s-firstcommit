package registry

import (
	"fmt"
	"k8s-firstcommit/pkg"
	"log"

	. "k8s-firstcommit/pkg/api"
)

func MakeEndpointController(serviceRegistry pkg.ServiceRegistry, taskRegistry pkg.TaskRegistry) *EndpointController {
	return &EndpointController{
		serviceRegistry: serviceRegistry,
		taskRegistry:    taskRegistry,
	}
}

type EndpointController struct {
	serviceRegistry pkg.ServiceRegistry
	taskRegistry    pkg.TaskRegistry
}

func (e *EndpointController) SyncServiceEndpoints() error {
	services, err := e.serviceRegistry.ListServices()
	if err != nil {
		return err
	}
	var resultErr error
	for _, service := range services.Items {
		tasks, err := e.taskRegistry.ListTasks(&service.Labels)
		if err != nil {
			log.Printf("Error syncing service: %#v, skipping.", service)
			resultErr = err
			continue
		}
		endpoints := make([]string, len(tasks))
		for ix, task := range tasks {
			// TODO: Use port names in the service object, don't just use port #0
			endpoints[ix] = fmt.Sprintf("%s:%d", task.CurrentState.Host, task.DesiredState.Manifest.Containers[0].Ports[0].HostPort)
		}
		err = e.serviceRegistry.UpdateEndpoints(Endpoints{
			Name:      service.ID,
			Endpoints: endpoints,
		})
		if err != nil {
			log.Printf("Error updating endpoints: %#v", err)
			continue
		}
	}
	return resultErr
}
