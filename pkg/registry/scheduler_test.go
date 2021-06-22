package registry

import (
	"math/rand"
	"testing"

	"k8s-firstcommit/pkg/api"
)

func expectSchedule(scheduler Scheduler, task api.Task, expected string, t *testing.T) {
	actual, err := scheduler.Schedule(task)
	expectNoError(t, err)
	if actual != expected {
		t.Errorf("Unexpected scheduling value: %d, expected %d", actual, expected)
	}
}

func TestRoundRobinScheduler(t *testing.T) {
	scheduler := MakeRoundRobinScheduler([]string{"m1", "m2", "m3", "m4"})
	expectSchedule(scheduler, api.Task{}, "m1", t)
	expectSchedule(scheduler, api.Task{}, "m2", t)
	expectSchedule(scheduler, api.Task{}, "m3", t)
	expectSchedule(scheduler, api.Task{}, "m4", t)
}

func TestRandomScheduler(t *testing.T) {
	random := rand.New(rand.NewSource(0))
	scheduler := MakeRandomScheduler([]string{"m1", "m2", "m3", "m4"}, *random)
	_, err := scheduler.Schedule(api.Task{})
	expectNoError(t, err)
}

func TestFirstFitSchedulerNothingScheduled(t *testing.T) {
	mockRegistry := MockTaskRegistry{}
	scheduler := MakeFirstFitScheduler([]string{"m1", "m2", "m3"}, &mockRegistry)
	expectSchedule(scheduler, api.Task{}, "m1", t)
}

func makeTask(host string, hostPorts ...int) api.Task {
	networkPorts := []api.Port{}
	for _, port := range hostPorts {
		networkPorts = append(networkPorts, api.Port{HostPort: port})
	}
	return api.Task{
		CurrentState: api.TaskState{
			Host: host,
		},
		DesiredState: api.TaskState{
			Manifest: api.ContainerManifest{
				Containers: []api.Container{
					api.Container{
						Ports: networkPorts,
					},
				},
			},
		},
	}
}

func TestFirstFitSchedulerFirstScheduled(t *testing.T) {
	mockRegistry := MockTaskRegistry{
		tasks: []api.Task{
			makeTask("m1", 8080),
		},
	}
	scheduler := MakeFirstFitScheduler([]string{"m1", "m2", "m3"}, &mockRegistry)
	expectSchedule(scheduler, makeTask("", 8080), "m2", t)
}

func TestFirstFitSchedulerFirstScheduledComplicated(t *testing.T) {
	mockRegistry := MockTaskRegistry{
		tasks: []api.Task{
			makeTask("m1", 80, 8080),
			makeTask("m2", 8081, 8082, 8083),
			makeTask("m3", 80, 443, 8085),
		},
	}
	scheduler := MakeFirstFitScheduler([]string{"m1", "m2", "m3"}, &mockRegistry)
	expectSchedule(scheduler, makeTask("", 8080, 8081), "m3", t)
}

func TestFirstFitSchedulerFirstScheduledImpossible(t *testing.T) {
	mockRegistry := MockTaskRegistry{
		tasks: []api.Task{
			makeTask("m1", 8080),
			makeTask("m2", 8081),
			makeTask("m3", 8080),
		},
	}
	scheduler := MakeFirstFitScheduler([]string{"m1", "m2", "m3"}, &mockRegistry)
	_, err := scheduler.Schedule(makeTask("", 8080, 8081))
	if err == nil {
		t.Error("Unexpected non-error.")
	}
}
