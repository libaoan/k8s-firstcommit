package registry

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"k8s-firstcommit/pkg/api"
	"k8s-firstcommit/pkg/apiserver"
)

type ServiceRegistry interface {
	ListServices() (api.ServiceList, error)
	CreateService(svc api.Service) error
	GetService(name string) (*api.Service, error)
	DeleteService(name string) error
	UpdateService(svc api.Service) error
	UpdateEndpoints(e api.Endpoints) error
}

type ServiceRegistryStorage struct {
	registry ServiceRegistry
}

func MakeServiceRegistryStorage(registry ServiceRegistry) apiserver.RESTStorage {
	return &ServiceRegistryStorage{registry: registry}
}

// GetServiceEnvironmentVariables populates a list of environment variables that are use
// in the container environment to get access to services.
func GetServiceEnvironmentVariables(registry ServiceRegistry, machine string) ([]api.EnvVar, error) {
	var result []api.EnvVar
	services, err := registry.ListServices()
	if err != nil {
		return result, err
	}
	for _, service := range services.Items {
		name := strings.ToUpper(service.ID) + "_SERVICE_PORT"
		value := strconv.Itoa(service.Port)
		result = append(result, api.EnvVar{Name: name, Value: value})
	}
	result = append(result, api.EnvVar{Name: "SERVICE_HOST", Value: machine})
	return result, nil
}

func (sr *ServiceRegistryStorage) List(*url.URL) (interface{}, error) {
	return sr.registry.ListServices()
}

func (sr *ServiceRegistryStorage) Get(id string) (interface{}, error) {
	return sr.registry.GetService(id)
}

func (sr *ServiceRegistryStorage) Delete(id string) error {
	return sr.registry.DeleteService(id)
}

func (sr *ServiceRegistryStorage) Extract(body string) (interface{}, error) {
	var svc api.Service
	err := json.Unmarshal([]byte(body), &svc)
	return svc, err
}

func (sr *ServiceRegistryStorage) Create(obj interface{}) error {
	return sr.registry.CreateService(obj.(api.Service))
}

func (sr *ServiceRegistryStorage) Update(obj interface{}) error {
	return sr.registry.UpdateService(obj.(api.Service))
}
