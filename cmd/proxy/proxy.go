package main

import (
	"flag"
	"log"
	"os"

	"github.com/coreos/go-etcd/etcd"
	"k8s-firstcommit/pkg/proxy"
	"k8s-firstcommit/pkg/proxy/config"
)

var (
	config_file  = flag.String("configfile", "/tmp/proxy_config", "Configuration file for the proxy")
	etcd_servers = flag.String("etcd_servers", "http://10.240.10.57:4001", "Servers for the etcd cluster (http://ip:port).")
)

func main() {
	flag.Parse()

	// Set up logger for etcd client
	etcd.SetLogger(log.New(os.Stderr, "etcd ", log.LstdFlags))

	log.Printf("Using configuration file %s and etcd_servers %s", *config_file, *etcd_servers)

	proxyConfig := config.NewServiceConfig()

	// Create a configuration source that handles configuration from etcd.
	etcdClient := etcd.NewClient([]string{*etcd_servers})
	config.NewConfigSourceEtcd(etcdClient,
		proxyConfig.GetServiceConfigurationChannel("etcd"),
		proxyConfig.GetEndpointsConfigurationChannel("etcd"))

	// And create a configuration source that reads from a local file
	config.NewConfigSourceFile(*config_file,
		proxyConfig.GetServiceConfigurationChannel("file"),
		proxyConfig.GetEndpointsConfigurationChannel("file"))

	loadBalancer := proxy.NewLoadBalancerRR()
	proxier := proxy.NewProxier(loadBalancer)
	// Wire proxier to handle changes to services
	proxyConfig.RegisterServiceHandler(proxier)
	// And wire loadBalancer to handle changes to endpoints to services
	proxyConfig.RegisterEndpointsHandler(loadBalancer)

	// Just loop forever for now...
	select {}

}
