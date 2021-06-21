package registry

import (
	"fmt"
	"k8s-firstcommit/pkg"
	"math/rand"

	. "k8s-firstcommit/pkg/api"
)

// Scheduler is an interface implemented by things that know how to schedule tasks onto machines.
type Scheduler interface {
	Schedule(Task) (string, error)
}

// RandomScheduler choses machines uniformly at random.
type RandomScheduler struct {
	machines []string
	random   rand.Rand
}

func MakeRandomScheduler(machines []string, random rand.Rand) Scheduler {
	return &RandomScheduler{
		machines: machines,
		random:   random,
	}
}

func (s *RandomScheduler) Schedule(task Task) (string, error) {
	return s.machines[s.random.Int()%len(s.machines)], nil
}

// RoundRobinScheduler chooses machines in order.
type RoundRobinScheduler struct {
	machines     []string
	currentIndex int
}

func MakeRoundRobinScheduler(machines []string) Scheduler {
	return &RoundRobinScheduler{
		machines:     machines,
		currentIndex: 0,
	}
}

func (s *RoundRobinScheduler) Schedule(task Task) (string, error) {
	result := s.machines[s.currentIndex]
	s.currentIndex = (s.currentIndex + 1) % len(s.machines)
	return result, nil
}

type FirstFitScheduler struct {
	machines []string
	registry pkg.TaskRegistry
}

func MakeFirstFitScheduler(machines []string, registry pkg.TaskRegistry) Scheduler {
	return &FirstFitScheduler{
		machines: machines,
		registry: registry,
	}
}

func (s *FirstFitScheduler) containsPort(task Task, port Port) bool {
	for _, container := range task.DesiredState.Manifest.Containers {
		for _, taskPort := range container.Ports {
			if taskPort.HostPort == port.HostPort {
				return true
			}
		}
	}
	return false
}

func (s *FirstFitScheduler) Schedule(task Task) (string, error) {
	machineToTasks := map[string][]Task{}
	tasks, err := s.registry.ListTasks(nil)
	if err != nil {
		return "", err
	}
	for _, scheduledTask := range tasks {
		host := scheduledTask.CurrentState.Host
		machineToTasks[host] = append(machineToTasks[host], scheduledTask)
	}
	for _, machine := range s.machines {
		taskFits := true
		for _, scheduledTask := range machineToTasks[machine] {
			for _, container := range task.DesiredState.Manifest.Containers {
				for _, port := range container.Ports {
					if s.containsPort(scheduledTask, port) {
						taskFits = false
					}
				}
			}
		}
		if taskFits {
			return machine, nil
		}
	}
	return "", fmt.Errorf("Failed to find fit for %#v", task)
}
