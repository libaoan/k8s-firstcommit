package registry

import (
	"k8s-firstcommit/pkg/api"
)

// An implementation of TaskRegistry and ControllerRegistry that is backed by memory
// Mainly used for testing.
type MemoryRegistry struct {
	taskData       map[string]api.Task
	controllerData map[string]api.ReplicationController
	serviceData    map[string]api.Service
}

func MakeMemoryRegistry() *MemoryRegistry {
	return &MemoryRegistry{
		taskData:       map[string]api.Task{},
		controllerData: map[string]api.ReplicationController{},
		serviceData:    map[string]api.Service{},
	}
}

func (registry *MemoryRegistry) ListTasks(labelQuery *map[string]string) ([]api.Task, error) {
	result := []api.Task{}
	for _, value := range registry.taskData {
		if LabelsMatch(value, labelQuery) {
			result = append(result, value)
		}
	}
	return result, nil
}

func (registry *MemoryRegistry) GetTask(taskID string) (*api.Task, error) {
	task, found := registry.taskData[taskID]
	if found {
		return &task, nil
	} else {
		return nil, nil
	}
}

func (registry *MemoryRegistry) CreateTask(machine string, task api.Task) error {
	registry.taskData[task.ID] = task
	return nil
}

func (registry *MemoryRegistry) DeleteTask(taskID string) error {
	delete(registry.taskData, taskID)
	return nil
}

func (registry *MemoryRegistry) UpdateTask(task api.Task) error {
	registry.taskData[task.ID] = task
	return nil
}

func (registry *MemoryRegistry) ListControllers() ([]api.ReplicationController, error) {
	result := []api.ReplicationController{}
	for _, value := range registry.controllerData {
		result = append(result, value)
	}
	return result, nil
}

func (registry *MemoryRegistry) GetController(controllerID string) (*api.ReplicationController, error) {
	controller, found := registry.controllerData[controllerID]
	if found {
		return &controller, nil
	} else {
		return nil, nil
	}
}

func (registry *MemoryRegistry) CreateController(controller api.ReplicationController) error {
	registry.controllerData[controller.ID] = controller
	return nil
}

func (registry *MemoryRegistry) DeleteController(controllerId string) error {
	delete(registry.controllerData, controllerId)
	return nil
}

func (registry *MemoryRegistry) UpdateController(controller api.ReplicationController) error {
	registry.controllerData[controller.ID] = controller
	return nil
}

func (registry *MemoryRegistry) ListServices() (api.ServiceList, error) {
	var list []api.Service
	for _, value := range registry.serviceData {
		list = append(list, value)
	}
	return api.ServiceList{Items: list}, nil
}

func (registry *MemoryRegistry) CreateService(svc api.Service) error {
	registry.serviceData[svc.ID] = svc
	return nil
}

func (registry *MemoryRegistry) GetService(name string) (*api.Service, error) {
	svc, found := registry.serviceData[name]
	if found {
		return &svc, nil
	} else {
		return nil, nil
	}
}

func (registry *MemoryRegistry) DeleteService(name string) error {
	delete(registry.serviceData, name)
	return nil
}

func (registry *MemoryRegistry) UpdateService(svc api.Service) error {
	return registry.CreateService(svc)
}

func (registry *MemoryRegistry) UpdateEndpoints(e api.Endpoints) error {
	return nil
}
