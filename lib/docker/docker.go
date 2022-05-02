package docker

import (
	"fmt"
	"strings"

	mp "github.com/mackerelio/go-mackerel-plugin"
	docker "github.com/fsouza/go-dockerclient"
)

type dockerClient interface {
	ListContainers(docker.ListContainersOptions) ([]docker.APIContainers, error)
}

type DockerPlugin struct {
	prefix string
	client dockerClient
}

func NewDockerPlugin(host, prefix string) (*DockerPlugin, error) {
	client, err := docker.NewClient(host)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}
	return &DockerPlugin{
		client: client,
		prefix: prefix,
	}, nil
}

func (m *DockerPlugin) listContainer() ([]docker.APIContainers, error) {
	containers, err := m.client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to execute listContainers API: %w", err)
	}

	return containers, nil
}

var statesMap map[string]struct{}

func init() {
	states := []string{
		"created",
		"restarting",
		"exited",
		"paused",
		"dead",
	}
	statesMap = make(map[string]struct{}, len(states))
	for _, s := range states {
		statesMap[s] = struct{}{}
	}
}

func (m *DockerPlugin) FetchMetrics() (map[string]float64, error) {
	containers, err := m.listContainer()
	if err != nil {
		return nil, fmt.Errorf("failed to get the list of containers: %w", err)
	}

	metrics := map[string]float64{}

	for _, c := range containers {
		if _, ok := statesMap[c.State]; ok {
			metrics[c.State] += 1
			continue
		}

		if c.State != MetricNameRunning {
			return nil, fmt.Errorf("met unknown state of docker container. state: %s", c.State)
		}
		status := getHealthCheckStatus(c.Status)
		switch status {
		case StatusStarting:
			metrics[MetricNameRunningStarting] += 1
		case StatusHealthy:
			metrics[MetricNameRunningHealthy] += 1
		case StatusUnhealthy:
			metrics[MetricNameRunningUnhealthy] += 1
		default:
			metrics[MetricNameRunning] += 1
		}
	}
	return metrics, nil
}

func (m *DockerPlugin) GraphDefinition() map[string]mp.Graphs {
	return map[string]mp.Graphs{
		"statuses": {
			Label: m.MetricKeyPrefix() + " Docker Container Statuses",
			Unit: mp.UnitInteger,
			Metrics: []mp.Metrics{
				{Name: "created", Label: "Created", Stacked: true},
				{Name: "running", Label: "Running", Stacked: true},
				{Name: "running_starting", Label: "Running (Health: starting)", Stacked: true},
				{Name: "running_unhealthy", Label: "Running (Health: unhealthy)", Stacked: true},
				{Name: "running_healthy", Label: "Running (Health: healthy)", Stacked: true},
				{Name: "restarting", Label: "Restarting", Stacked: true},
				{Name: "exited", Label: "Exited", Stacked: true},
				{Name: "paused", Label: "Paused", Stacked: true},
				{Name: "dead", Label: "Dead", Stacked: true},
			},
		},
	}
}
func (m *DockerPlugin) MetricKeyPrefix() string {
	return m.prefix
}

type Status int

const (
	StatusUnknown Status = iota
	StatusStarting
	StatusHealthy
	StatusUnhealthy
)

const (
	StatusSuffixStarting  = "(health: starting)"
	StatusSuffixHealthy   = "(healthy)"
	StatusSuffixUnhealthy = "(unhealthy)"

	MetricNameRunning = "running"
	MetricNameRunningStarting  = "running_starting"
	MetricNameRunningHealthy   = "running_healthy"
	MetricNameRunningUnhealthy = "running_unhealthy"
)

func getHealthCheckStatus(status string) Status {
	if strings.HasSuffix(status, StatusSuffixUnhealthy) {
		return StatusUnhealthy
	}

	if strings.HasSuffix(status, StatusSuffixStarting) {
		return StatusStarting
	}

	if strings.HasSuffix(status, StatusSuffixHealthy) {
		return StatusHealthy
	}

	return StatusUnknown
}
