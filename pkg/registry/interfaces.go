package registry

import (
	"k8s-firstcommit/pkg/api"
)

// TaskRegistry is an interface implemented by things that know how to store Task objects
type TaskRegistry interface {
	// ListTasks obtains a list of tasks that match query.
	// Query may be nil in which case all tasks are returned.
	ListTasks(query *map[string]string) ([]api.Task, error)
	// Get a specific task
	GetTask(taskId string) (*api.Task, error)
	// Create a task based on a specification, schedule it onto a specific machine.
	CreateTask(machine string, task api.Task) error
	// Update an existing task
	UpdateTask(task api.Task) error
	// Delete an existing task
	DeleteTask(taskId string) error
}

// ControllerRegistry is an interface for things that know how to store Controllers
type ControllerRegistry interface {
	ListControllers() ([]api.ReplicationController, error)
	GetController(controllerId string) (*api.ReplicationController, error)
	CreateController(controller api.ReplicationController) error
	UpdateController(controller api.ReplicationController) error
	DeleteController(controllerId string) error
}
