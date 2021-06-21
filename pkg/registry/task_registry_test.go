package registry

import (
	"encoding/json"
	"fmt"
	"k8s-firstcommit/pkg"
	"testing"

	. "k8s-firstcommit/pkg/api"
)

type MockTaskRegistry struct {
	err   error
	tasks []Task
}

func expectNoError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
}

func (registry *MockTaskRegistry) ListTasks(*map[string]string) ([]Task, error) {
	return registry.tasks, registry.err
}

func (registry *MockTaskRegistry) GetTask(taskId string) (*Task, error) {
	return &Task{}, registry.err
}

func (registry *MockTaskRegistry) CreateTask(machine string, task Task) error {
	return registry.err
}

func (registry *MockTaskRegistry) UpdateTask(task Task) error {
	return registry.err
}
func (registry *MockTaskRegistry) DeleteTask(taskId string) error {
	return registry.err
}

func TestListTasksError(t *testing.T) {
	mockRegistry := MockTaskRegistry{
		err: fmt.Errorf("Test Error"),
	}
	storage := pkg.TaskRegistryStorage{
		registry: &mockRegistry,
	}
	tasks, err := storage.List(nil)
	if err != mockRegistry.err {
		t.Errorf("Expected %#v, Got %#v", mockRegistry.err, err)
	}
	if len(tasks.(TaskList).Items) != 0 {
		t.Errorf("Unexpected non-zero task list: %#v", tasks)
	}
}

func TestListEmptyTaskList(t *testing.T) {
	mockRegistry := MockTaskRegistry{}
	storage := pkg.TaskRegistryStorage{
		registry: &mockRegistry,
	}
	tasks, err := storage.List(nil)
	expectNoError(t, err)
	if len(tasks.(TaskList).Items) != 0 {
		t.Errorf("Unexpected non-zero task list: %#v", tasks)
	}
}

func TestListTaskList(t *testing.T) {
	mockRegistry := MockTaskRegistry{
		tasks: []Task{
			Task{
				JSONBase: JSONBase{
					ID: "foo",
				},
			},
			Task{
				JSONBase: JSONBase{
					ID: "bar",
				},
			},
		},
	}
	storage := pkg.TaskRegistryStorage{
		registry: &mockRegistry,
	}
	tasksObj, err := storage.List(nil)
	tasks := tasksObj.(TaskList)
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
	storage := pkg.TaskRegistryStorage{
		registry: &mockRegistry,
	}
	task := Task{
		JSONBase: JSONBase{
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

func expectLabelMatch(t *testing.T, task Task, key, value string) {
	if !pkg.LabelMatch(task, key, value) {
		t.Errorf("Unexpected match failure: %#v %s %s", task, key, value)
	}
}

func expectNoLabelMatch(t *testing.T, task Task, key, value string) {
	if pkg.LabelMatch(task, key, value) {
		t.Errorf("Unexpected match success: %#v %s %s", task, key, value)
	}
}

func expectLabelsMatch(t *testing.T, task Task, query *map[string]string) {
	if !pkg.LabelsMatch(task, query) {
		t.Errorf("Unexpected match failure: %#v %#v", task, *query)
	}
}

func expectNoLabelsMatch(t *testing.T, task Task, query *map[string]string) {
	if pkg.LabelsMatch(task, query) {
		t.Errorf("Unexpected match success: %#v %#v", task, *query)
	}
}

func TestLabelMatch(t *testing.T) {
	task := Task{
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
	task := Task{
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