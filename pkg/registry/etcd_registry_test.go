package registry

import (
	"encoding/json"
	"k8s-firstcommit/pkg"
	"reflect"
	"testing"

	"github.com/coreos/go-etcd/etcd"
	. "k8s-firstcommit/pkg/api"
	"k8s-firstcommit/pkg/util"
)

func TestEtcdGetTask(t *testing.T) {
	fakeClient := pkg.MakeFakeEtcdClient(t)
	fakeClient.Set("/registry/hosts/machine/tasks/foo", util.MakeJSONString(Task{JSONBase: JSONBase{ID: "foo"}}), 0)
	registry := pkg.MakeTestEtcdRegistry(fakeClient, []string{"machine"})
	task, err := registry.GetTask("foo")
	pkg.expectNoError(t, err)
	if task.ID != "foo" {
		t.Errorf("Unexpected task: %#v", task)
	}
}

func TestEtcdGetTaskNotFound(t *testing.T) {
	fakeClient := pkg.MakeFakeEtcdClient(t)
	fakeClient.Data["/registry/hosts/machine/tasks/foo"] = pkg.EtcdResponseWithError{
		R: &etcd.Response{
			Node: nil,
		},
		E: &etcd.EtcdError{
			ErrorCode: 100,
		},
	}
	registry := pkg.MakeTestEtcdRegistry(fakeClient, []string{"machine"})
	_, err := registry.GetTask("foo")
	if err == nil {
		t.Errorf("Unexpected non-error.")
	}
}

func TestEtcdCreateTask(t *testing.T) {
	fakeClient := pkg.MakeFakeEtcdClient(t)
	fakeClient.Data["/registry/hosts/machine/tasks/foo"] = pkg.EtcdResponseWithError{
		R: &etcd.Response{
			Node: nil,
		},
		E: &etcd.EtcdError{ErrorCode: 100},
	}
	fakeClient.Set("/registry/hosts/machine/kubelet", util.MakeJSONString([]ContainerManifest{}), 0)
	registry := pkg.MakeTestEtcdRegistry(fakeClient, []string{"machine"})
	err := registry.CreateTask("machine", Task{
		JSONBase: JSONBase{
			ID: "foo",
		},
		DesiredState: TaskState{
			Manifest: ContainerManifest{
				Containers: []Container{
					Container{
						Name: "foo",
					},
				},
			},
		},
	})
	pkg.expectNoError(t, err)
	resp, err := fakeClient.Get("/registry/hosts/machine/tasks/foo", false, false)
	pkg.expectNoError(t, err)
	var task Task
	err = json.Unmarshal([]byte(resp.Node.Value), &task)
	pkg.expectNoError(t, err)
	if task.ID != "foo" {
		t.Errorf("Unexpected task: %#v %s", task, resp.Node.Value)
	}
	var manifests []ContainerManifest
	resp, err = fakeClient.Get("/registry/hosts/machine/kubelet", false, false)
	pkg.expectNoError(t, err)
	err = json.Unmarshal([]byte(resp.Node.Value), &manifests)
	if len(manifests) != 1 || manifests[0].Id != "foo" {
		t.Errorf("Unexpected manifest list: %#v", manifests)
	}
}

func TestEtcdCreateTaskAlreadyExisting(t *testing.T) {
	fakeClient := pkg.MakeFakeEtcdClient(t)
	fakeClient.Data["/registry/hosts/machine/tasks/foo"] = pkg.EtcdResponseWithError{
		R: &etcd.Response{
			Node: &etcd.Node{
				Value: util.MakeJSONString(Task{JSONBase: JSONBase{ID: "foo"}}),
			},
		},
		E: nil,
	}
	registry := pkg.MakeTestEtcdRegistry(fakeClient, []string{"machine"})
	err := registry.CreateTask("machine", Task{
		JSONBase: JSONBase{
			ID: "foo",
		},
	})
	if err == nil {
		t.Error("Unexpected non-error")
	}
}

func TestEtcdCreateTaskWithContainersError(t *testing.T) {
	fakeClient := pkg.MakeFakeEtcdClient(t)
	fakeClient.Data["/registry/hosts/machine/tasks/foo"] = pkg.EtcdResponseWithError{
		R: &etcd.Response{
			Node: nil,
		},
		E: &etcd.EtcdError{ErrorCode: 100},
	}
	fakeClient.Data["/registry/hosts/machine/kubelet"] = pkg.EtcdResponseWithError{
		R: &etcd.Response{
			Node: nil,
		},
		E: &etcd.EtcdError{ErrorCode: 200},
	}
	registry := pkg.MakeTestEtcdRegistry(fakeClient, []string{"machine"})
	err := registry.CreateTask("machine", Task{
		JSONBase: JSONBase{
			ID: "foo",
		},
	})
	if err == nil {
		t.Error("Unexpected non-error")
	}
	_, err = fakeClient.Get("/registry/hosts/machine/tasks/foo", false, false)
	if err == nil {
		t.Error("Unexpected non-error")
	}
	if err != nil && err.(*etcd.EtcdError).ErrorCode != 100 {
		t.Errorf("Unexpected error: %#v", err)
	}
}

func TestEtcdCreateTaskWithContainersNotFound(t *testing.T) {
	fakeClient := pkg.MakeFakeEtcdClient(t)
	fakeClient.Data["/registry/hosts/machine/tasks/foo"] = pkg.EtcdResponseWithError{
		R: &etcd.Response{
			Node: nil,
		},
		E: &etcd.EtcdError{ErrorCode: 100},
	}
	fakeClient.Data["/registry/hosts/machine/kubelet"] = pkg.EtcdResponseWithError{
		R: &etcd.Response{
			Node: nil,
		},
		E: &etcd.EtcdError{ErrorCode: 100},
	}
	registry := pkg.MakeTestEtcdRegistry(fakeClient, []string{"machine"})
	err := registry.CreateTask("machine", Task{
		JSONBase: JSONBase{
			ID: "foo",
		},
		DesiredState: TaskState{
			Manifest: ContainerManifest{
				Id: "foo",
				Containers: []Container{
					Container{
						Name: "foo",
					},
				},
			},
		},
	})
	pkg.expectNoError(t, err)
	resp, err := fakeClient.Get("/registry/hosts/machine/tasks/foo", false, false)
	pkg.expectNoError(t, err)
	var task Task
	err = json.Unmarshal([]byte(resp.Node.Value), &task)
	pkg.expectNoError(t, err)
	if task.ID != "foo" {
		t.Errorf("Unexpected task: %#v %s", task, resp.Node.Value)
	}
	var manifests []ContainerManifest
	resp, err = fakeClient.Get("/registry/hosts/machine/kubelet", false, false)
	pkg.expectNoError(t, err)
	err = json.Unmarshal([]byte(resp.Node.Value), &manifests)
	if len(manifests) != 1 || manifests[0].Id != "foo" {
		t.Errorf("Unexpected manifest list: %#v", manifests)
	}
}

func TestEtcdCreateTaskWithExistingContainers(t *testing.T) {
	fakeClient := pkg.MakeFakeEtcdClient(t)
	fakeClient.Data["/registry/hosts/machine/tasks/foo"] = pkg.EtcdResponseWithError{
		R: &etcd.Response{
			Node: nil,
		},
		E: &etcd.EtcdError{ErrorCode: 100},
	}
	fakeClient.Set("/registry/hosts/machine/kubelet", util.MakeJSONString([]ContainerManifest{
		ContainerManifest{
			Id: "bar",
		},
	}), 0)
	registry := pkg.MakeTestEtcdRegistry(fakeClient, []string{"machine"})
	err := registry.CreateTask("machine", Task{
		JSONBase: JSONBase{
			ID: "foo",
		},
		DesiredState: TaskState{
			Manifest: ContainerManifest{
				Id: "foo",
				Containers: []Container{
					Container{
						Name: "foo",
					},
				},
			},
		},
	})
	pkg.expectNoError(t, err)
	resp, err := fakeClient.Get("/registry/hosts/machine/tasks/foo", false, false)
	pkg.expectNoError(t, err)
	var task Task
	err = json.Unmarshal([]byte(resp.Node.Value), &task)
	pkg.expectNoError(t, err)
	if task.ID != "foo" {
		t.Errorf("Unexpected task: %#v %s", task, resp.Node.Value)
	}
	var manifests []ContainerManifest
	resp, err = fakeClient.Get("/registry/hosts/machine/kubelet", false, false)
	pkg.expectNoError(t, err)
	err = json.Unmarshal([]byte(resp.Node.Value), &manifests)
	if len(manifests) != 2 || manifests[1].Id != "foo" {
		t.Errorf("Unexpected manifest list: %#v", manifests)
	}
}

func TestEtcdDeleteTask(t *testing.T) {
	fakeClient := pkg.MakeFakeEtcdClient(t)
	key := "/registry/hosts/machine/tasks/foo"
	fakeClient.Set(key, util.MakeJSONString(Task{JSONBase: JSONBase{ID: "foo"}}), 0)
	fakeClient.Set("/registry/hosts/machine/kubelet", util.MakeJSONString([]ContainerManifest{
		ContainerManifest{
			Id: "foo",
		},
	}), 0)
	registry := pkg.MakeTestEtcdRegistry(fakeClient, []string{"machine"})
	err := registry.DeleteTask("foo")
	pkg.expectNoError(t, err)
	if len(fakeClient.deletedKeys) != 1 {
		t.Errorf("Expected 1 delete, found %#v", fakeClient.deletedKeys)
	}
	if fakeClient.deletedKeys[0] != key {
		t.Errorf("Unexpected key: %s, expected %s", fakeClient.deletedKeys[0], key)
	}
	response, _ := fakeClient.Get("/registry/hosts/machine/kubelet", false, false)
	if response.Node.Value != "[]" {
		t.Errorf("Unexpected container set: %s, expected empty", response.Node.Value)
	}
}

func TestEtcdDeleteTaskMultipleContainers(t *testing.T) {
	fakeClient := pkg.MakeFakeEtcdClient(t)
	key := "/registry/hosts/machine/tasks/foo"
	fakeClient.Set(key, util.MakeJSONString(Task{JSONBase: JSONBase{ID: "foo"}}), 0)
	fakeClient.Set("/registry/hosts/machine/kubelet", util.MakeJSONString([]ContainerManifest{
		ContainerManifest{Id: "foo"},
		ContainerManifest{Id: "bar"},
	}), 0)
	registry := pkg.MakeTestEtcdRegistry(fakeClient, []string{"machine"})
	err := registry.DeleteTask("foo")
	pkg.expectNoError(t, err)
	if len(fakeClient.deletedKeys) != 1 {
		t.Errorf("Expected 1 delete, found %#v", fakeClient.deletedKeys)
	}
	if fakeClient.deletedKeys[0] != key {
		t.Errorf("Unexpected key: %s, expected %s", fakeClient.deletedKeys[0], key)
	}
	response, _ := fakeClient.Get("/registry/hosts/machine/kubelet", false, false)
	var manifests []ContainerManifest
	json.Unmarshal([]byte(response.Node.Value), &manifests)
	if len(manifests) != 1 {
		t.Errorf("Unexpected manifest set: %#v, expected empty", manifests)
	}
	if manifests[0].Id != "bar" {
		t.Errorf("Deleted wrong manifest: %#v", manifests)
	}
}

func TestEtcdEmptyListTasks(t *testing.T) {
	fakeClient := pkg.MakeFakeEtcdClient(t)
	key := "/registry/hosts/machine/tasks"
	fakeClient.Data[key] = pkg.EtcdResponseWithError{
		R: &etcd.Response{
			Node: &etcd.Node{
				Nodes: []*etcd.Node{},
			},
		},
		E: nil,
	}
	registry := pkg.MakeTestEtcdRegistry(fakeClient, []string{"machine"})
	tasks, err := registry.ListTasks(nil)
	pkg.expectNoError(t, err)
	if len(tasks) != 0 {
		t.Errorf("Unexpected task list: %#v", tasks)
	}
}

func TestEtcdListTasksNotFound(t *testing.T) {
	fakeClient := pkg.MakeFakeEtcdClient(t)
	key := "/registry/hosts/machine/tasks"
	fakeClient.Data[key] = pkg.EtcdResponseWithError{
		R: &etcd.Response{},
		E: &etcd.EtcdError{ErrorCode: 100},
	}
	registry := pkg.MakeTestEtcdRegistry(fakeClient, []string{"machine"})
	tasks, err := registry.ListTasks(nil)
	pkg.expectNoError(t, err)
	if len(tasks) != 0 {
		t.Errorf("Unexpected task list: %#v", tasks)
	}
}

func TestEtcdListTasks(t *testing.T) {
	fakeClient := pkg.MakeFakeEtcdClient(t)
	key := "/registry/hosts/machine/tasks"
	fakeClient.Data[key] = pkg.EtcdResponseWithError{
		R: &etcd.Response{
			Node: &etcd.Node{
				Nodes: []*etcd.Node{
					&etcd.Node{
						Value: util.MakeJSONString(Task{JSONBase: JSONBase{ID: "foo"}}),
					},
					&etcd.Node{
						Value: util.MakeJSONString(Task{JSONBase: JSONBase{ID: "bar"}}),
					},
				},
			},
		},
		E: nil,
	}
	registry := pkg.MakeTestEtcdRegistry(fakeClient, []string{"machine"})
	tasks, err := registry.ListTasks(nil)
	pkg.expectNoError(t, err)
	if len(tasks) != 2 || tasks[0].ID != "foo" || tasks[1].ID != "bar" {
		t.Errorf("Unexpected task list: %#v", tasks)
	}
}

func TestEtcdListControllersNotFound(t *testing.T) {
	fakeClient := pkg.MakeFakeEtcdClient(t)
	key := "/registry/controllers"
	fakeClient.Data[key] = pkg.EtcdResponseWithError{
		R: &etcd.Response{},
		E: &etcd.EtcdError{ErrorCode: 100},
	}
	registry := pkg.MakeTestEtcdRegistry(fakeClient, []string{"machine"})
	controllers, err := registry.ListControllers()
	pkg.expectNoError(t, err)
	if len(controllers) != 0 {
		t.Errorf("Unexpected controller list: %#v", controllers)
	}
}

func TestEtcdListServicesNotFound(t *testing.T) {
	fakeClient := pkg.MakeFakeEtcdClient(t)
	key := "/registry/services/specs"
	fakeClient.Data[key] = pkg.EtcdResponseWithError{
		R: &etcd.Response{},
		E: &etcd.EtcdError{ErrorCode: 100},
	}
	registry := pkg.MakeTestEtcdRegistry(fakeClient, []string{"machine"})
	services, err := registry.ListServices()
	pkg.expectNoError(t, err)
	if len(services.Items) != 0 {
		t.Errorf("Unexpected controller list: %#v", services)
	}
}

func TestEtcdListControllers(t *testing.T) {
	fakeClient := pkg.MakeFakeEtcdClient(t)
	key := "/registry/controllers"
	fakeClient.Data[key] = pkg.EtcdResponseWithError{
		R: &etcd.Response{
			Node: &etcd.Node{
				Nodes: []*etcd.Node{
					&etcd.Node{
						Value: util.MakeJSONString(ReplicationController{JSONBase: JSONBase{ID: "foo"}}),
					},
					&etcd.Node{
						Value: util.MakeJSONString(ReplicationController{JSONBase: JSONBase{ID: "bar"}}),
					},
				},
			},
		},
		E: nil,
	}
	registry := pkg.MakeTestEtcdRegistry(fakeClient, []string{"machine"})
	controllers, err := registry.ListControllers()
	pkg.expectNoError(t, err)
	if len(controllers) != 2 || controllers[0].ID != "foo" || controllers[1].ID != "bar" {
		t.Errorf("Unexpected controller list: %#v", controllers)
	}
}

func TestEtcdGetController(t *testing.T) {
	fakeClient := pkg.MakeFakeEtcdClient(t)
	fakeClient.Set("/registry/controllers/foo", util.MakeJSONString(ReplicationController{JSONBase: JSONBase{ID: "foo"}}), 0)
	registry := pkg.MakeTestEtcdRegistry(fakeClient, []string{"machine"})
	ctrl, err := registry.GetController("foo")
	pkg.expectNoError(t, err)
	if ctrl.ID != "foo" {
		t.Errorf("Unexpected controller: %#v", ctrl)
	}
}

func TestEtcdGetControllerNotFound(t *testing.T) {
	fakeClient := pkg.MakeFakeEtcdClient(t)
	fakeClient.Data["/registry/controllers/foo"] = pkg.EtcdResponseWithError{
		R: &etcd.Response{
			Node: nil,
		},
		E: &etcd.EtcdError{
			ErrorCode: 100,
		},
	}
	registry := pkg.MakeTestEtcdRegistry(fakeClient, []string{"machine"})
	ctrl, err := registry.GetController("foo")
	if ctrl != nil {
		t.Errorf("Unexpected non-nil controller: %#v", ctrl)
	}
	if err == nil {
		t.Error("Unexpected non-error.")
	}
}

func TestEtcdDeleteController(t *testing.T) {
	fakeClient := pkg.MakeFakeEtcdClient(t)
	registry := pkg.MakeTestEtcdRegistry(fakeClient, []string{"machine"})
	err := registry.DeleteController("foo")
	pkg.expectNoError(t, err)
	if len(fakeClient.deletedKeys) != 1 {
		t.Errorf("Expected 1 delete, found %#v", fakeClient.deletedKeys)
	}
	key := "/registry/controllers/foo"
	if fakeClient.deletedKeys[0] != key {
		t.Errorf("Unexpected key: %s, expected %s", fakeClient.deletedKeys[0], key)
	}
}

func TestEtcdCreateController(t *testing.T) {
	fakeClient := pkg.MakeFakeEtcdClient(t)
	registry := pkg.MakeTestEtcdRegistry(fakeClient, []string{"machine"})
	err := registry.CreateController(ReplicationController{
		JSONBase: JSONBase{
			ID: "foo",
		},
	})
	pkg.expectNoError(t, err)
	resp, err := fakeClient.Get("/registry/controllers/foo", false, false)
	pkg.expectNoError(t, err)
	var ctrl ReplicationController
	err = json.Unmarshal([]byte(resp.Node.Value), &ctrl)
	pkg.expectNoError(t, err)
	if ctrl.ID != "foo" {
		t.Errorf("Unexpected task: %#v %s", ctrl, resp.Node.Value)
	}
}

func TestEtcdUpdateController(t *testing.T) {
	fakeClient := pkg.MakeFakeEtcdClient(t)
	fakeClient.Set("/registry/controllers/foo", util.MakeJSONString(ReplicationController{JSONBase: JSONBase{ID: "foo"}}), 0)
	registry := pkg.MakeTestEtcdRegistry(fakeClient, []string{"machine"})
	err := registry.UpdateController(ReplicationController{
		JSONBase: JSONBase{ID: "foo"},
		DesiredState: ReplicationControllerState{
			Replicas: 2,
		},
	})
	pkg.expectNoError(t, err)
	ctrl, err := registry.GetController("foo")
	if ctrl.DesiredState.Replicas != 2 {
		t.Errorf("Unexpected controller: %#v", ctrl)
	}
}

func TestEtcdListServices(t *testing.T) {
	fakeClient := pkg.MakeFakeEtcdClient(t)
	key := "/registry/services/specs"
	fakeClient.Data[key] = pkg.EtcdResponseWithError{
		R: &etcd.Response{
			Node: &etcd.Node{
				Nodes: []*etcd.Node{
					&etcd.Node{
						Value: util.MakeJSONString(Service{JSONBase: JSONBase{ID: "foo"}}),
					},
					&etcd.Node{
						Value: util.MakeJSONString(Service{JSONBase: JSONBase{ID: "bar"}}),
					},
				},
			},
		},
		E: nil,
	}
	registry := pkg.MakeTestEtcdRegistry(fakeClient, []string{"machine"})
	services, err := registry.ListServices()
	pkg.expectNoError(t, err)
	if len(services.Items) != 2 || services.Items[0].ID != "foo" || services.Items[1].ID != "bar" {
		t.Errorf("Unexpected task list: %#v", services)
	}
}

func TestEtcdCreateService(t *testing.T) {
	fakeClient := pkg.MakeFakeEtcdClient(t)
	fakeClient.Data["/registry/services/specs/foo"] = pkg.EtcdResponseWithError{
		R: &etcd.Response{
			Node: nil,
		},
		E: &etcd.EtcdError{ErrorCode: 100},
	}
	registry := pkg.MakeTestEtcdRegistry(fakeClient, []string{"machine"})
	err := registry.CreateService(Service{
		JSONBase: JSONBase{ID: "foo"},
	})
	pkg.expectNoError(t, err)
	resp, err := fakeClient.Get("/registry/services/specs/foo", false, false)
	pkg.expectNoError(t, err)
	var service Service
	err = json.Unmarshal([]byte(resp.Node.Value), &service)
	pkg.expectNoError(t, err)
	if service.ID != "foo" {
		t.Errorf("Unexpected service: %#v %s", service, resp.Node.Value)
	}
}

func TestEtcdGetService(t *testing.T) {
	fakeClient := pkg.MakeFakeEtcdClient(t)
	fakeClient.Set("/registry/services/specs/foo", util.MakeJSONString(Service{JSONBase: JSONBase{ID: "foo"}}), 0)
	registry := pkg.MakeTestEtcdRegistry(fakeClient, []string{"machine"})
	service, err := registry.GetService("foo")
	pkg.expectNoError(t, err)
	if service.ID != "foo" {
		t.Errorf("Unexpected task: %#v", service)
	}
}

func TestEtcdGetServiceNotFound(t *testing.T) {
	fakeClient := pkg.MakeFakeEtcdClient(t)
	fakeClient.Data["/registry/services/specs/foo"] = pkg.EtcdResponseWithError{
		R: &etcd.Response{
			Node: nil,
		},
		E: &etcd.EtcdError{
			ErrorCode: 100,
		},
	}
	registry := pkg.MakeTestEtcdRegistry(fakeClient, []string{"machine"})
	_, err := registry.GetService("foo")
	if err == nil {
		t.Errorf("Unexpected non-error.")
	}
}

func TestEtcdDeleteService(t *testing.T) {
	fakeClient := pkg.MakeFakeEtcdClient(t)
	registry := pkg.MakeTestEtcdRegistry(fakeClient, []string{"machine"})
	err := registry.DeleteService("foo")
	pkg.expectNoError(t, err)
	if len(fakeClient.deletedKeys) != 2 {
		t.Errorf("Expected 2 delete, found %#v", fakeClient.deletedKeys)
	}
	key := "/registry/services/specs/foo"
	if fakeClient.deletedKeys[0] != key {
		t.Errorf("Unexpected key: %s, expected %s", fakeClient.deletedKeys[0], key)
	}
	key = "/registry/services/endpoints/foo"
	if fakeClient.deletedKeys[1] != key {
		t.Errorf("Unexpected key: %s, expected %s", fakeClient.deletedKeys[1], key)
	}
}

func TestEtcdUpdateService(t *testing.T) {
	fakeClient := pkg.MakeFakeEtcdClient(t)
	fakeClient.Set("/registry/services/specs/foo", util.MakeJSONString(Service{JSONBase: JSONBase{ID: "foo"}}), 0)
	registry := pkg.MakeTestEtcdRegistry(fakeClient, []string{"machine"})
	err := registry.UpdateService(Service{
		JSONBase: JSONBase{ID: "foo"},
		Labels: map[string]string{
			"baz": "bar",
		},
	})
	pkg.expectNoError(t, err)
	svc, err := registry.GetService("foo")
	if svc.Labels["baz"] != "bar" {
		t.Errorf("Unexpected service: %#v", svc)
	}
}

func TestEtcdUpdateEndpoints(t *testing.T) {
	fakeClient := pkg.MakeFakeEtcdClient(t)
	registry := pkg.MakeTestEtcdRegistry(fakeClient, []string{"machine"})
	endpoints := Endpoints{
		Name:      "foo",
		Endpoints: []string{"baz", "bar"},
	}
	err := registry.UpdateEndpoints(endpoints)
	pkg.expectNoError(t, err)
	response, err := fakeClient.Get("/registry/services/endpoints/foo", false, false)
	pkg.expectNoError(t, err)
	var endpointsOut Endpoints
	err = json.Unmarshal([]byte(response.Node.Value), &endpointsOut)
	if !reflect.DeepEqual(endpoints, endpointsOut) {
		t.Errorf("Unexpected endpoints: %#v, expected %#v", endpointsOut, endpoints)
	}
}
