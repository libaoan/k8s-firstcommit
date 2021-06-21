package registry

import (
	"k8s-firstcommit/pkg"
	"testing"

	. "k8s-firstcommit/pkg/api"
)

func TestListTasksEmpty(t *testing.T) {
	registry := pkg.MakeMemoryRegistry()
	tasks, err := registry.ListTasks(nil)
	pkg.expectNoError(t, err)
	if len(tasks) != 0 {
		t.Errorf("Unexpected task list: %#v", tasks)
	}
}

func TestMemoryListTasks(t *testing.T) {
	registry := pkg.MakeMemoryRegistry()
	registry.CreateTask("machine", Task{JSONBase: JSONBase{ID: "foo"}})
	tasks, err := registry.ListTasks(nil)
	pkg.expectNoError(t, err)
	if len(tasks) != 1 || tasks[0].ID != "foo" {
		t.Errorf("Unexpected task list: %#v", tasks)
	}
}

func TestMemorySetGetTasks(t *testing.T) {
	registry := pkg.MakeMemoryRegistry()
	expectedTask := Task{JSONBase: JSONBase{ID: "foo"}}
	registry.CreateTask("machine", expectedTask)
	task, err := registry.GetTask("foo")
	pkg.expectNoError(t, err)
	if expectedTask.ID != task.ID {
		t.Errorf("Unexpected task, expected %#v, actual %#v", expectedTask, task)
	}
}

func TestMemorySetUpdateGetTasks(t *testing.T) {
	registry := pkg.MakeMemoryRegistry()
	oldTask := Task{JSONBase: JSONBase{ID: "foo"}}
	expectedTask := Task{
		JSONBase: JSONBase{
			ID: "foo",
		},
		DesiredState: TaskState{
			Host: "foo.com",
		},
	}
	registry.CreateTask("machine", oldTask)
	registry.UpdateTask(expectedTask)
	task, err := registry.GetTask("foo")
	pkg.expectNoError(t, err)
	if expectedTask.ID != task.ID || task.DesiredState.Host != expectedTask.DesiredState.Host {
		t.Errorf("Unexpected task, expected %#v, actual %#v", expectedTask, task)
	}
}

func TestMemorySetDeleteGetTasks(t *testing.T) {
	registry := pkg.MakeMemoryRegistry()
	expectedTask := Task{JSONBase: JSONBase{ID: "foo"}}
	registry.CreateTask("machine", expectedTask)
	registry.DeleteTask("foo")
	task, err := registry.GetTask("foo")
	pkg.expectNoError(t, err)
	if task != nil {
		t.Errorf("Unexpected task: %#v", task)
	}
}

func TestListControllersEmpty(t *testing.T) {
	registry := pkg.MakeMemoryRegistry()
	tasks, err := registry.ListControllers()
	pkg.expectNoError(t, err)
	if len(tasks) != 0 {
		t.Errorf("Unexpected task list: %#v", tasks)
	}
}

func TestMemoryListControllers(t *testing.T) {
	registry := pkg.MakeMemoryRegistry()
	registry.CreateController(ReplicationController{JSONBase: JSONBase{ID: "foo"}})
	tasks, err := registry.ListControllers()
	pkg.expectNoError(t, err)
	if len(tasks) != 1 || tasks[0].ID != "foo" {
		t.Errorf("Unexpected task list: %#v", tasks)
	}
}

func TestMemorySetGetControllers(t *testing.T) {
	registry := pkg.MakeMemoryRegistry()
	expectedController := ReplicationController{JSONBase: JSONBase{ID: "foo"}}
	registry.CreateController(expectedController)
	task, err := registry.GetController("foo")
	pkg.expectNoError(t, err)
	if expectedController.ID != task.ID {
		t.Errorf("Unexpected task, expected %#v, actual %#v", expectedController, task)
	}
}

func TestMemorySetUpdateGetControllers(t *testing.T) {
	registry := pkg.MakeMemoryRegistry()
	oldController := ReplicationController{JSONBase: JSONBase{ID: "foo"}}
	expectedController := ReplicationController{
		JSONBase: JSONBase{
			ID: "foo",
		},
		DesiredState: ReplicationControllerState{
			Replicas: 2,
		},
	}
	registry.CreateController(oldController)
	registry.UpdateController(expectedController)
	task, err := registry.GetController("foo")
	pkg.expectNoError(t, err)
	if expectedController.ID != task.ID || task.DesiredState.Replicas != expectedController.DesiredState.Replicas {
		t.Errorf("Unexpected task, expected %#v, actual %#v", expectedController, task)
	}
}

func TestMemorySetDeleteGetControllers(t *testing.T) {
	registry := pkg.MakeMemoryRegistry()
	expectedController := ReplicationController{JSONBase: JSONBase{ID: "foo"}}
	registry.CreateController(expectedController)
	registry.DeleteController("foo")
	task, err := registry.GetController("foo")
	pkg.expectNoError(t, err)
	if task != nil {
		t.Errorf("Unexpected task: %#v", task)
	}
}
