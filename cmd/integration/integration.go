// A basic integration test for the service.
// Assumes that there is a pre-existing etcd server running on localhost.
package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http/httptest"
	"time"

	"github.com/coreos/go-etcd/etcd"
	"k8s-firstcommit/pkg/api"
	"k8s-firstcommit/pkg/apiserver"
	kube_client "k8s-firstcommit/pkg/client"
	"k8s-firstcommit/pkg/registry"
)

func main() {

	// Setup
	servers := []string{"http://localhost:4001"}
	log.Printf("Creating etcd client pointing to %v", servers)
	etcdClient := etcd.NewClient(servers)
	machineList := []string{"machine"}

	reg := registry.MakeEtcdRegistry(etcdClient, machineList)

	apiserver := apiserver.New(map[string]apiserver.RESTStorage{
		"tasks":                  registry.MakeTaskRegistryStorage(reg, &kube_client.FakeContainerInfo{}, registry.MakeRoundRobinScheduler(machineList)),
		"replicationControllers": registry.MakeControllerRegistryStorage(reg),
	}, "/api/v1beta1")
	server := httptest.NewServer(apiserver)

	controllerManager := registry.MakeReplicationManager(etcd.NewClient(servers),
		kube_client.Client{
			Host: server.URL,
		})

	go controllerManager.Synchronize()
	go controllerManager.WatchControllers()

	// Ok. we're good to go.
	log.Printf("API Server started on %s", server.URL)
	// Wait for the synchronization threads to come up.
	time.Sleep(time.Second * 10)

	kubeClient := kube_client.Client{
		Host: server.URL,
	}
	data, err := ioutil.ReadFile("api/examples/controller.json")
	if err != nil {
		log.Fatalf("Unexpected error: %#v", err)
	}
	var controllerRequest api.ReplicationController
	if err = json.Unmarshal(data, &controllerRequest); err != nil {
		log.Fatalf("Unexpected error: %#v", err)
	}

	if _, err = kubeClient.CreateReplicationController(controllerRequest); err != nil {
		log.Fatalf("Unexpected error: %#v", err)
	}
	// Give the controllers some time to actually create the tasks
	time.Sleep(time.Second * 10)

	// Validate that they're truly up.
	tasks, err := kubeClient.ListTasks(nil)
	if err != nil || len(tasks.Items) != 2 {
		log.Fatal("FAILED")
	}
	log.Printf("OK")
}
