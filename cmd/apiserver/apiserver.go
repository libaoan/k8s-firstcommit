package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/coreos/go-etcd/etcd"

	"k8s-firstcommit/pkg/apiserver"
	kube_client "k8s-firstcommit/pkg/client"
	"k8s-firstcommit/pkg/registry"
	"k8s-firstcommit/pkg/util"
)

var (
	port                        = flag.Uint("port", 8080, "The port to listen on.  Default 8080.")
	address                     = flag.String("address", "127.0.0.1", "The address on the local server to listen to. Default 127.0.0.1")
	apiPrefix                   = flag.String("api_prefix", "/api/v1beta1", "The prefix for API requests on the server. Default '/api/v1beta1'")
	etcdServerList, machineList util.StringList
)

func init() {
	flag.Var(&etcdServerList, "etcd_servers", "Servers for the etcd (http://ip:port), comma separated")
	flag.Var(&machineList, "machines", "List of machines to schedule onto, comma separated.")
}

func main() {
	flag.Parse()

	if len(machineList) == 0 {
		log.Fatal("No machines specified!")
	}

	var (
		taskRegistry       registry.TaskRegistry
		controllerRegistry registry.ControllerRegistry
		serviceRegistry    registry.ServiceRegistry
	)

	if len(etcdServerList) > 0 {
		log.Printf("Creating etcd client pointing to %v", etcdServerList)
		etcdClient := etcd.NewClient(etcdServerList)
		taskRegistry = registry.MakeEtcdRegistry(etcdClient, machineList)
		controllerRegistry = registry.MakeEtcdRegistry(etcdClient, machineList)
		serviceRegistry = registry.MakeEtcdRegistry(etcdClient, machineList)
	} else {
		taskRegistry = registry.MakeMemoryRegistry()
		controllerRegistry = registry.MakeMemoryRegistry()
		serviceRegistry = registry.MakeMemoryRegistry()
	}

	containerInfo := &kube_client.HTTPContainerInfo{
		Client: http.DefaultClient,
		Port:   10250,
	}

	storage := map[string]apiserver.RESTStorage{
		"tasks":                  registry.MakeTaskRegistryStorage(taskRegistry, containerInfo, registry.MakeFirstFitScheduler(machineList, taskRegistry)),
		"replicationControllers": registry.MakeControllerRegistryStorage(controllerRegistry),
		"services":               registry.MakeServiceRegistryStorage(serviceRegistry),
	}

	endpoints := registry.MakeEndpointController(serviceRegistry, taskRegistry)
	go util.Forever(func() { endpoints.SyncServiceEndpoints() }, time.Second*10)

	s := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", *address, *port),
		Handler:        apiserver.New(storage, *apiPrefix),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
