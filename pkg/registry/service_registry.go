package registry

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	. "k8s-firstcommit/pkg/api"
	"k8s-firstcommit/pkg/apiserver"
)

type ServiceRegistry interface {
	ListServices() (ServiceList, error)
	CreateService(svc Service) error
	GetService(name string) (*Service, error)
	DeleteService(name string) error
	UpdateService(svc Service) error
	UpdateEndpoints(e Endpoints) error
}

type ServiceRegistryStorage struct {
	registry ServiceRegistry
}

func MakeServiceRegistryStorage(registry ServiceRegistry) apiserver.RESTStorage {
	return &ServiceRegistryStorage{registry: registry}
}

// GetServiceEnvironmentVariables populates a list of environment variables that are use
// in the container environment to get access to services.
func GetServiceEnvironmentVariables(registry ServiceRegistry, machine string) ([]EnvVar, error) {
	var result []EnvVar
	services, err := registry.ListServices()
	if err != nil {
		return result, err
	}
	for _, service := range services.Items {
		name := strings.ToUpper(service.ID) + "_SERVICE_PORT"
		value := strconv.Itoa(service.Port)
		result = append(result, EnvVar{Name: name, Value: value})
	}
	result = append(result, EnvVar{Name: "SERVICE_HOST", Value: machine})
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
	var svc Service
	err := json.Unmarshal([]byte(body), &svc)
	return svc, err
}

func (sr *ServiceRegistryStorage) Create(obj interface{}) error {
	return sr.registry.CreateService(obj.(Service))
}

func (sr *ServiceRegistryStorage) Update(obj interface{}) error {
	return sr.registry.UpdateService(obj.(Service))
}
