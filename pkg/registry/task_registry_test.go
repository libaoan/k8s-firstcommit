package registry

import (
	"encoding/json"
	"fmt"
	"testing"

	"k8s-firstcommit/pkg/api"
)

type MockTaskRegistry struct {
	err   error
	tasks []api.Task
}

func expectNoError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
}

func (registry *MockTaskRegistry) ListTasks(*map[string]string) ([]api.Task, error) {
	return registry.tasks, registry.err
}

func (registry *MockTaskRegistry) GetTask(taskId string) (*api.Task, error) {
	return &api.Task{}, registry.err
}

func (registry *MockTaskRegistry) CreateTask(machine string, task api.Task) error {
	return registry.err
}

func (registry *MockTaskRegistry) UpdateTask(task api.Task) error {
	return registry.err
}
func (registry *MockTaskRegistry) DeleteTask(taskId string) error {
	return registry.err
}

func TestListTasksError(t *testing.T) {
	mockRegistry := MockTaskRegistry{
		err: fmt.Errorf("Test Error"),
	}
	storage := TaskRegistryStorage{
		registry: &mockRegistry,
	}
	tasks, err := storage.List(nil)
	if err != mockRegistry.err {
		t.Errorf("Expected %#v, Got %#v", mockRegistry.err, err)
	}
	if len(tasks.(api.TaskList).Items) != 0 {
		t.Errorf("Unexpected non-zero task list: %#v", tasks)
	}
}

func TestListEmptyTaskList(t *testing.T) {
	mockRegistry := MockTaskRegistry{}
	storage := TaskRegistryStorage{
		registry: &mockRegistry,
	}
	tasks, err := storage.List(nil)
	expectNoError(t, err)
	if len(tasks.(api.TaskList).Items) != 0 {
		t.Errorf("Unexpected non-zero task list: %#v", tasks)
	}
}

func TestListTaskList(t *testing.T) {
	mockRegistry := MockTaskRegistry{
		tasks: []api.Task{
			api.Task{
				JSONBase: api.JSONBase{
					ID: "foo",
				},
			},
			api.Task{
				JSONBase: api.JSONBase{
					ID: "bar",
				},
			},
		},
	}
	storage := TaskRegistryStorage{
		registry: &mockRegistry,
	}
	tasksObj, err := storage.List(nil)
	tasks := tasksObj.(api.TaskList)
	expectNoError(t, err)
	if len(tasks.Items) != 2 {
		t.Errorf("Unexpected task list: %#v", tasks)
	}
	if tasks.Items[0].ID != "foo" {
		t.Errorf("Unexpected task: %#v", tasks.Items[0])
	}
	if tasks.Items[1].ID != "bar" {
		t.Errorf("Unexpected task: %#v", tasks.Items[1])
	}
}

func TestExtractJson(t *testing.T) {
	mockRegistry := MockTaskRegistry{}
	storage := TaskRegistryStorage{
		registry: &mockRegistry,
	}
	task := api.Task{
		JSONBase: api.JSONBase{
			ID: "foo",
		},
	}
	body, err := json.Marshal(task)
	expectNoError(t, err)
	taskOut, err := storage.Extract(string(body))
	expectNoError(t, err)
	jsonOut, err := json.Marshal(taskOut)
	expectNoError(t, err)
	if string(body) != string(jsonOut) {
		t.Errorf("Expected %#v, found %#v", task, taskOut)
	}
}

func expectLabelMatch(t *testing.T, task api.Task, key, value string) {
	if !LabelMatch(task, key, value) {
		t.Errorf("Unexpected match failure: %#v %s %s", task, key, value)
	}
}

func expectNoLabelMatch(t *testing.T, task api.Task, key, value string) {
	if LabelMatch(task, key, value) {
		t.Errorf("Unexpected match success: %#v %s %s", task, key, value)
	}
}

func expectLabelsMatch(t *testing.T, task api.Task, query *map[string]string) {
	if !LabelsMatch(task, query) {
		t.Errorf("Unexpected match failure: %#v %#v", task, *query)
	}
}

func expectNoLabelsMatch(t *testing.T, task api.Task, query *map[string]string) {
	if LabelsMatch(task, query) {
		t.Errorf("Unexpected match success: %#v %#v", task, *query)
	}
}

func TestLabelMatch(t *testing.T) {
	task := api.Task{
		Labels: map[string]string{
			"foo": "bar",
			"baz": "blah",
		},
	}
	expectLabelMatch(t, task, "foo", "bar")
	expectLabelMatch(t, task, "baz", "blah")
	expectNoLabelMatch(t, task, "foo", "blah")
	expectNoLabelMatch(t, task, "baz", "bar")
}

func TestLabelsMatch(t *testing.T) {
	task := api.Task{
		Labels: map[string]string{
			"foo": "bar",
			"baz": "blah",
		},
	}
	expectLabelsMatch(t, task, &map[string]string{})
	expectLabelsMatch(t, task, &map[string]string{
		"foo": "bar",
	})
	expectLabelsMatch(t, task, &map[string]string{
		"baz": "blah",
	})
	expectLabelsMatch(t, task, &map[string]string{
		"foo": "bar",
		"baz": "blah",
	})
	expectNoLabelsMatch(t, task, &map[string]string{
		"foo": "blah",
	})
	expectNoLabelsMatch(t, task, &map[string]string{
		"baz": "bar",
	})
	expectNoLabelsMatch(t, task, &map[string]string{
		"foo":    "bar",
		"foobar": "bar",
		"baz":    "blah",
	})

}
