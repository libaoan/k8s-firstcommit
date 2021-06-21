package registry

import (
	"k8s-firstcommit/pkg"
	"math/rand"
	"testing"

	. "k8s-firstcommit/pkg/api"
)

func expectSchedule(scheduler pkg.Scheduler, task Task, expected string, t *testing.T) {
	actual, err := scheduler.Schedule(task)
	pkg.expectNoError(t, err)
	if actual != expected {
		t.Errorf("Unexpected scheduling value: %d, expected %d", actual, expected)
	}
}

func TestRoundRobinScheduler(t *testing.T) {
	scheduler := pkg.MakeRoundRobinScheduler([]string{"m1", "m2", "m3", "m4"})
	expectSchedule(scheduler, Task{}, "m1", t)
	expectSchedule(scheduler, Task{}, "m2", t)
	expectSchedule(scheduler, Task{}, "m3", t)
	expectSchedule(scheduler, Task{}, "m4", t)
}

func TestRandomScheduler(t *testing.T) {
	random := rand.New(rand.NewSource(0))
	scheduler := pkg.MakeRandomScheduler([]string{"m1", "m2", "m3", "m4"}, *random)
	_, err := scheduler.Schedule(Task{})
	pkg.expectNoError(t, err)
}

func TestFirstFitSchedulerNothingScheduled(t *testing.T) {
	mockRegistry := pkg.MockTaskRegistry{}
	scheduler := pkg.MakeFirstFitScheduler([]string{"m1", "m2", "m3"}, &mockRegistry)
	expectSchedule(scheduler, Task{}, "m1", t)
}

func makeTask(host string, hostPorts ...int) Task {
	networkPorts := []Port{}
	for _, port := range hostPorts {
		networkPorts = append(networkPorts, Port{HostPort: port})
	}
	return Task{
		CurrentState: TaskState{
			Host: host,
		},
		DesiredState: TaskState{
			Manifest: ContainerManifest{
				Containers: []Container{
					Container{
						Ports: networkPorts,
					},
				},
			},
		},
	}
}

func TestFirstFitSchedulerFirstScheduled(t *testing.T) {
	mockRegistry := pkg.MockTaskRegistry{
		tasks: []Task{
			makeTask("m1", 8080),
		},
	}
	scheduler := pkg.MakeFirstFitScheduler([]string{"m1", "m2", "m3"}, &mockRegistry)
	expectSchedule(scheduler, makeTask("", 8080), "m2", t)
}

func TestFirstFitSchedulerFirstScheduledComplicated(t *testing.T) {
	mockRegistry := pkg.MockTaskRegistry{
		tasks: []Task{
			makeTask("m1", 80, 8080),
			makeTask("m2", 8081, 8082, 8083),
			makeTask("m3", 80, 443, 8085),
		},
	}
	scheduler := pkg.MakeFirstFitScheduler([]string{"m1", "m2", "m3"}, &mockRegistry)
	expectSchedule(scheduler, makeTask("", 8080, 8081), "m3", t)
}

func TestFirstFitSchedulerFirstScheduledImpossible(t *testing.T) {
	mockRegistry := pkg.MockTaskRegistry{
		tasks: []Task{
			makeTask("m1", 8080),
			makeTask("m2", 8081),
			makeTask("m3", 8080),
		},
	}
	scheduler := pkg.MakeFirstFitScheduler([]string{"m1", "m2", "m3"}, &mockRegistry)
	_, err := scheduler.Schedule(makeTask("", 8080, 8081))
	if err == nil {
		t.Error("Unexpected non-error.")
	}
}
