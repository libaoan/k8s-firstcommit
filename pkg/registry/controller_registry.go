package registry

import (
	"encoding/json"

	"net/url"

	"k8s-firstcommit/pkg/api"
	"k8s-firstcommit/pkg/apiserver"
)

// Implementation of RESTStorage for the api server.
type ControllerRegistryStorage struct {
	registry ControllerRegistry
}

func MakeControllerRegistryStorage(registry ControllerRegistry) apiserver.RESTStorage {
	return &ControllerRegistryStorage{
		registry: registry,
	}
}

func (storage *ControllerRegistryStorage) List(*url.URL) (interface{}, error) {
	var result api.ReplicationControllerList
	controllers, err := storage.registry.ListControllers()
	if err == nil {
		result = api.ReplicationControllerList{
			Items: controllers,
		}
	}
	return result, err
}

func (storage *ControllerRegistryStorage) Get(id string) (interface{}, error) {
	return storage.registry.GetController(id)
}

func (storage *ControllerRegistryStorage) Delete(id string) error {
	return storage.registry.DeleteController(id)
}

func (storage *ControllerRegistryStorage) Extract(body string) (interface{}, error) {
	result := api.ReplicationController{}
	err := json.Unmarshal([]byte(body), &result)
	return result, err
}

func (storage *ControllerRegistryStorage) Create(controller interface{}) error {
	return storage.registry.CreateController(controller.(api.ReplicationController))
}

func (storage *ControllerRegistryStorage) Update(controller interface{}) error {
	return storage.registry.UpdateController(controller.(api.ReplicationController))
}
