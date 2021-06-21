package registry

import (
	"encoding/json"
	"k8s-firstcommit/pkg"
	"net/url"

	. "k8s-firstcommit/pkg/api"
	"k8s-firstcommit/pkg/apiserver"
)

// Implementation of RESTStorage for the api server.
type ControllerRegistryStorage struct {
	registry pkg.ControllerRegistry
}

func MakeControllerRegistryStorage(registry pkg.ControllerRegistry) apiserver.RESTStorage {
	return &ControllerRegistryStorage{
		registry: registry,
	}
}

func (storage *ControllerRegistryStorage) List(*url.URL) (interface{}, error) {
	var result ReplicationControllerList
	controllers, err := storage.registry.ListControllers()
	if err == nil {
		result = ReplicationControllerList{
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
	result := ReplicationController{}
	err := json.Unmarshal([]byte(body), &result)
	return result, err
}

func (storage *ControllerRegistryStorage) Create(controller interface{}) error {
	return storage.registry.CreateController(controller.(ReplicationController))
}

func (storage *ControllerRegistryStorage) Update(controller interface{}) error {
	return storage.registry.UpdateController(controller.(ReplicationController))
}
