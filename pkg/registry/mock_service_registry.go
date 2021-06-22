package registry

import (
	"k8s-firstcommit/pkg/api"
)

type MockServiceRegistry struct {
	list      api.ServiceList
	err       error
	endpoints api.Endpoints
}

func (m *MockServiceRegistry) ListServices() (api.ServiceList, error) {
	return m.list, m.err
}

func (m *MockServiceRegistry) CreateService(svc api.Service) error {
	return m.err
}

func (m *MockServiceRegistry) GetService(name string) (*api.Service, error) {
	return nil, m.err
}

func (m *MockServiceRegistry) DeleteService(name string) error {
	return m.err
}

func (m *MockServiceRegistry) UpdateService(svc api.Service) error {
	return m.err
}

func (m *MockServiceRegistry) UpdateEndpoints(e api.Endpoints) error {
	m.endpoints = e
	return m.err
}
