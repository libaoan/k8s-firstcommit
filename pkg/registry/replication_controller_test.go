package registry

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/coreos/go-etcd/etcd"
	"k8s-firstcommit/pkg/api"
	"k8s-firstcommit/pkg/client"
	"k8s-firstcommit/pkg/util"
)

// TODO: Move this to a common place, it's needed in multiple tests.
var apiPath = "/api/v1beta1"

func makeUrl(suffix string) string {
	return apiPath + suffix
}

type FakeTaskControl struct {
	controllerSpec []api.ReplicationController
	deleteTaskID   []string
}

func (f *FakeTaskControl) createReplica(spec api.ReplicationController) {
	f.controllerSpec = append(f.controllerSpec, spec)
}

func (f *FakeTaskControl) deleteTask(taskID string) error {
	f.deleteTaskID = append(f.deleteTaskID, taskID)
	return nil
}

func makeReplicationController(replicas int) api.ReplicationController {
	return api.ReplicationController{
		DesiredState: api.ReplicationControllerState{
			Replicas: replicas,
			TaskTemplate: api.TaskTemplate{
				DesiredState: api.TaskState{
					Manifest: api.ContainerManifest{
						Containers: []api.Container{
							api.Container{
								Image: "foo/bar",
							},
						},
					},
				},
				Labels: map[string]string{
					"name": "foo",
					"type": "production",
				},
			},
		},
	}
}

func makeTaskList(count int) api.TaskList {
	tasks := []api.Task{}
	for i := 0; i < count; i++ {
		tasks = append(tasks, api.Task{
			JSONBase: api.JSONBase{
				ID: fmt.Sprintf("task%d", i),
			},
		})
	}
	return api.TaskList{
		Items: tasks,
	}
}

func validateSyncReplication(t *testing.T, fakeTaskControl *FakeTaskControl, expectedCreates, expectedDeletes int) {
	if len(fakeTaskControl.controllerSpec) != expectedCreates {
		t.Errorf("Unexpected number of creates.  Expected %d, saw %d\n", expectedCreates, len(fakeTaskControl.controllerSpec))
	}
	if len(fakeTaskControl.deleteTaskID) != expectedDeletes {
		t.Errorf("Unexpected number of deletes.  Expected %d, saw %d\n", expectedDeletes, len(fakeTaskControl.deleteTaskID))
	}
}

func TestSyncReplicationControllerDoesNothing(t *testing.T) {
	body, _ := json.Marshal(makeTaskList(2))
	fakeHandler := util.FakeHandler{
		StatusCode:   200,
		ResponseBody: string(body),
	}
	testServer := httptest.NewTLSServer(&fakeHandler)
	client := client.Client{
		Host: testServer.URL,
	}

	fakeTaskControl := FakeTaskControl{}

	manager := MakeReplicationManager(nil, &client)
	manager.taskControl = &fakeTaskControl

	controllerSpec := makeReplicationController(2)

	manager.syncReplicationController(controllerSpec)
	validateSyncReplication(t, &fakeTaskControl, 0, 0)
}

func TestSyncReplicationControllerDeletes(t *testing.T) {
	body, _ := json.Marshal(makeTaskList(2))
	fakeHandler := util.FakeHandler{
		StatusCode:   200,
		ResponseBody: string(body),
	}
	testServer := httptest.NewTLSServer(&fakeHandler)
	client := client.Client{
		Host: testServer.URL,
	}

	fakeTaskControl := FakeTaskControl{}

	manager := MakeReplicationManager(nil, &client)
	manager.taskControl = &fakeTaskControl

	controllerSpec := makeReplicationController(1)

	manager.syncReplicationController(controllerSpec)
	validateSyncReplication(t, &fakeTaskControl, 0, 1)
}

func TestSyncReplicationControllerCreates(t *testing.T) {
	body := "{ \"items\": [] }"
	fakeHandler := util.FakeHandler{
		StatusCode:   200,
		ResponseBody: string(body),
	}
	testServer := httptest.NewTLSServer(&fakeHandler)
	client := client.Client{
		Host: testServer.URL,
	}

	fakeTaskControl := FakeTaskControl{}

	manager := MakeReplicationManager(nil, &client)
	manager.taskControl = &fakeTaskControl

	controllerSpec := makeReplicationController(2)

	manager.syncReplicationController(controllerSpec)
	validateSyncReplication(t, &fakeTaskControl, 2, 0)
}

func TestCreateReplica(t *testing.T) {
	body := "{}"
	fakeHandler := util.FakeHandler{
		StatusCode:   200,
		ResponseBody: string(body),
	}
	testServer := httptest.NewTLSServer(&fakeHandler)
	client := client.Client{
		Host: testServer.URL,
	}

	taskControl := RealTaskControl{
		kubeClient: client,
	}

	controllerSpec := api.ReplicationController{
		DesiredState: api.ReplicationControllerState{
			TaskTemplate: api.TaskTemplate{
				DesiredState: api.TaskState{
					Manifest: api.ContainerManifest{
						Containers: []api.Container{
							api.Container{
								Image: "foo/bar",
							},
						},
					},
				},
				Labels: map[string]string{
					"name": "foo",
					"type": "production",
				},
			},
		},
	}

	taskControl.createReplica(controllerSpec)

	//expectedTask := Task{
	//	Labels:       controllerSpec.DesiredState.TaskTemplate.Labels,
	//	DesiredState: controllerSpec.DesiredState.TaskTemplate.DesiredState,
	//}
	// TODO: fix this so that it validates the body.
	fakeHandler.ValidateRequest(t, makeUrl("/tasks"), "POST", nil)
}

func TestHandleWatchResponseNotSet(t *testing.T) {
	body, _ := json.Marshal(makeTaskList(2))
	fakeHandler := util.FakeHandler{
		StatusCode:   200,
		ResponseBody: string(body),
	}
	testServer := httptest.NewTLSServer(&fakeHandler)
	client := client.Client{
		Host: testServer.URL,
	}

	fakeTaskControl := FakeTaskControl{}

	manager := MakeReplicationManager(nil, &client)
	manager.taskControl = &fakeTaskControl
	_, err := manager.handleWatchResponse(&etcd.Response{
		Action: "delete",
	})
	expectNoError(t, err)
}

func TestHandleWatchResponseNoNode(t *testing.T) {
	body, _ := json.Marshal(makeTaskList(2))
	fakeHandler := util.FakeHandler{
		StatusCode:   200,
		ResponseBody: string(body),
	}
	testServer := httptest.NewTLSServer(&fakeHandler)
	client := client.Client{
		Host: testServer.URL,
	}

	fakeTaskControl := FakeTaskControl{}

	manager := MakeReplicationManager(nil, &client)
	manager.taskControl = &fakeTaskControl
	_, err := manager.handleWatchResponse(&etcd.Response{
		Action: "set",
	})
	if err == nil {
		t.Error("Unexpected non-error")
	}
}

func TestHandleWatchResponseBadData(t *testing.T) {
	body, _ := json.Marshal(makeTaskList(2))
	fakeHandler := util.FakeHandler{
		StatusCode:   200,
		ResponseBody: string(body),
	}
	testServer := httptest.NewTLSServer(&fakeHandler)
	client := client.Client{
		Host: testServer.URL,
	}

	fakeTaskControl := FakeTaskControl{}

	manager := MakeReplicationManager(nil, &client)
	manager.taskControl = &fakeTaskControl
	_, err := manager.handleWatchResponse(&etcd.Response{
		Action: "set",
		Node: &etcd.Node{
			Value: "foobar",
		},
	})
	if err == nil {
		t.Error("Unexpected non-error")
	}
}

func TestHandleWatchResponse(t *testing.T) {
	body, _ := json.Marshal(makeTaskList(2))
	fakeHandler := util.FakeHandler{
		StatusCode:   200,
		ResponseBody: string(body),
	}
	testServer := httptest.NewTLSServer(&fakeHandler)
	client := client.Client{
		Host: testServer.URL,
	}

	fakeTaskControl := FakeTaskControl{}

	manager := MakeReplicationManager(nil, &client)
	manager.taskControl = &fakeTaskControl

	controller := makeReplicationController(2)

	data, err := json.Marshal(controller)
	expectNoError(t, err)
	controllerOut, err := manager.handleWatchResponse(&etcd.Response{
		Action: "set",
		Node: &etcd.Node{
			Value: string(data),
		},
	})
	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
	if !reflect.DeepEqual(controller, *controllerOut) {
		t.Errorf("Unexpected mismatch.  Expected %#v, Saw: %#v", controller, controllerOut)
	}
}
