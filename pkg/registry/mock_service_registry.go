package registry

import (
	. "k8s-firstcommit/pkg/api"
)

type MockServiceRegistry struct {
	list      ServiceList
	err       error
	endpoints Endpoints
}

func (m *MockServiceRegistry) ListServices() (ServiceList, error) {
	return m.list, m.err
}

func (m *MockServiceRegistry) CreateService(svc Service) error {
	return m.err
}

func (m *MockServiceRegistry) GetService(name string) (*Service, error) {
	return nil, m.err
}

func (m *MockServiceRegistry) DeleteService(name string) error {
	return m.err
}

func (m *MockServiceRegistry) UpdateService(svc Service) error {
	return m.err
}

func (m *MockServiceRegistry) UpdateEndpoints(e Endpoints) error {
	m.endpoints = e
	return m.err
}
