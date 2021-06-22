package registry

import (
	"fmt"
	"testing"

	"github.com/coreos/go-etcd/etcd"
)

type EtcdResponseWithError struct {
	R *etcd.Response
	E error
}

type FakeEtcdClient struct {
	Data        map[string]EtcdResponseWithError
	deletedKeys []string
	err         error
	t           *testing.T
}

func MakeFakeEtcdClient(t *testing.T) *FakeEtcdClient {
	return &FakeEtcdClient{
		t:    t,
		Data: map[string]EtcdResponseWithError{},
	}
}

func (f *FakeEtcdClient) AddChild(key, data string, ttl uint64) (*etcd.Response, error) {
	return f.Set(key, data, ttl)
}

func (f *FakeEtcdClient) Get(key string, sort, recursive bool) (*etcd.Response, error) {
	result := f.Data[key]
	if result.R == nil {
		f.t.Errorf("Unexpected get for %s", key)
		return &etcd.Response{}, &etcd.EtcdError{ErrorCode: 100}
	}
	return result.R, result.E
}

func (f *FakeEtcdClient) Set(key, value string, ttl uint64) (*etcd.Response, error) {
	result := EtcdResponseWithError{
		R: &etcd.Response{
			Node: &etcd.Node{
				Value: value,
			},
		},
	}
	f.Data[key] = result
	return result.R, f.err
}
func (f *FakeEtcdClient) Create(key, value string, ttl uint64) (*etcd.Response, error) {
	return f.Set(key, value, ttl)
}
func (f *FakeEtcdClient) Delete(key string, recursive bool) (*etcd.Response, error) {
	f.deletedKeys = append(f.deletedKeys, key)
	return &etcd.Response{}, f.err
}

func (f *FakeEtcdClient) Watch(prefix string, waitIndex uint64, recursive bool, receiver chan *etcd.Response, stop chan bool) (*etcd.Response, error) {
	return nil, fmt.Errorf("Unimplemented")
}

func MakeTestEtcdRegistry(client EtcdClient, machines []string) *EtcdRegistry {
	registry := MakeEtcdRegistry(client, machines)
	registry.manifestFactory = &BasicManifestFactory{
		serviceRegistry: &MockServiceRegistry{},
	}
	return registry
}
