package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"reflect"
	"time"

	"k8s-firstcommit/pkg/api"
)

// TODO: kill this struct.
type ServiceJSON struct {
	Name      string
	Port      int
	Endpoints []string
}
type ConfigFile struct {
	Services []ServiceJSON
}

type ConfigSourceFile struct {
	serviceChannel   chan ServiceUpdate
	endpointsChannel chan EndpointsUpdate
	filename         string
}

func NewConfigSourceFile(filename string, serviceChannel chan ServiceUpdate, endpointsChannel chan EndpointsUpdate) ConfigSourceFile {
	config := ConfigSourceFile{
		filename:         filename,
		serviceChannel:   serviceChannel,
		endpointsChannel: endpointsChannel,
	}
	go config.Run()
	return config
}

func (impl ConfigSourceFile) Run() {
	log.Printf("Watching file %s", impl.filename)
	var lastData []byte
	var lastServices []api.Service
	var lastEndpoints []api.Endpoints

	for {
		data, err := ioutil.ReadFile(impl.filename)
		if err != nil {
			log.Printf("Couldn't read file: %s : %v", impl.filename, err)
		} else {
			var config ConfigFile
			err = json.Unmarshal(data, &config)
			if err != nil {
				log.Printf("Couldn't unmarshal configuration from file : %s %v", data, err)
			} else {
				if !bytes.Equal(lastData, data) {
					lastData = data
					// Ok, we have a valid configuration, send to channel for
					// rejiggering.
					newServices := make([]api.Service, len(config.Services))
					newEndpoints := make([]api.Endpoints, len(config.Services))
					for i, service := range config.Services {
						newServices[i] = api.Service{JSONBase: api.JSONBase{ID: service.Name}, Port: service.Port}
						newEndpoints[i] = api.Endpoints{Name: service.Name, Endpoints: service.Endpoints}
					}
					if !reflect.DeepEqual(lastServices, newServices) {
						serviceUpdate := ServiceUpdate{Op: SET, Services: newServices}
						impl.serviceChannel <- serviceUpdate
						lastServices = newServices
					}
					if !reflect.DeepEqual(lastEndpoints, newEndpoints) {
						endpointsUpdate := EndpointsUpdate{Op: SET, Endpoints: newEndpoints}
						impl.endpointsChannel <- endpointsUpdate
						lastEndpoints = newEndpoints
					}
				}
			}
		}
		time.Sleep(5 * time.Second)
	}
}
