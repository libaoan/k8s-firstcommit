// TODO: Refactor the etcd watch code so that it is a pluggable interface.
package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/coreos/go-etcd/etcd"
	kube_client "k8s-firstcommit/pkg/client"
	"k8s-firstcommit/pkg/registry"
	"k8s-firstcommit/pkg/util"
)

var (
	etcd_servers = flag.String("etcd_servers", "", "Servers for the etcd (http://ip:port).")
	master       = flag.String("master", "", "The address of the Kubernetes API server")
)

func main() {
	flag.Parse()

	if len(*etcd_servers) == 0 || len(*master) == 0 {
		log.Fatal("usage: controller-manager -etcd_servers <servers> -master <master>")
	}

	// Set up logger for etcd client
	etcd.SetLogger(log.New(os.Stderr, "etcd ", log.LstdFlags))

	controllerManager := registry.MakeReplicationManager(etcd.NewClient([]string{*etcd_servers}),
		kube_client.Client{
			Host: "http://" + *master,
		})

	go util.Forever(func() { controllerManager.Synchronize() }, 20*time.Second)
	go util.Forever(func() { controllerManager.WatchControllers() }, 20*time.Second)
	select {}
}
