package docker

import (
	"testing"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/stretchr/testify/assert"
)

type mockDockerClient struct{}

func (m *mockDockerClient) ListContainers(_ docker.ListContainersOptions) ([]docker.APIContainers, error) {
	containers := []docker.APIContainers{
		docker.APIContainers{
			State:      "running",
		},
		docker.APIContainers{
			State:      "running",
			Status:     "Up 16 minutes (health: starting)",
		},
		docker.APIContainers{
			State:      "running",
			Status:     "Up 15 minutes (unhealthy)",
		},
		docker.APIContainers{
			State:      "running",
			Status:     "Up 15 minutes (healthy)",
		},
		docker.APIContainers{
			State:      "created",
		},
		docker.APIContainers{
			State:      "restarting",
		},
		docker.APIContainers{
			State:      "exited",
		},
		docker.APIContainers{
			State:      "paused",
		},
		docker.APIContainers{
			State:      "dead",
		},
	}
	return append(containers, containers...), nil
}

func TestFetchMetics(t *testing.T) {
	p := &DockerPlugin{
		client: &mockDockerClient{},
		prefix: "",
	}

	got, err := p.FetchMetrics()
	assert.NoError(t, err)
	want := map[string]float64{
		"created": 2,
		"running": 2,
		"running_starting": 2,
		"running_unhealthy": 2,
		"running_healthy": 2,
		"restarting": 2,
		"exited": 2,
		"paused": 2,
		"dead": 2,
	}
	assert.Equal(t, want, got)
}