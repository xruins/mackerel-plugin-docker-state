package docker

import (
	"fmt"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
	mp "github.com/mackerelio/go-mackerel-plugin"
)

type dockerClient interface {
	ListContainers(docker.ListContainersOptions) ([]docker.APIContainers, error)
}

type DockerPlugin struct {
	prefix           string
	client           dockerClient
	enableTotal      bool
	enableFailing    bool
	failingStatusMap map[string]struct{}
}

func NewDockerPlugin(host, prefix string, enableTotal, enableFailing bool, failingStatuses []string) (*DockerPlugin, error) {
	client, err := docker.NewClient(host)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	failingStatusMap := make(map[string]struct{}, len(failingStatuses))
	for _, s := range failingStatuses {
		var found bool
		for _, ps := range allStatuses {
			if s == ps {
				found = true
				break
			}
		}
		if found {
			failingStatusMap[s] = struct{}{}
			continue
		}
		return nil, fmt.Errorf("invalid failingStatuses. status: %s", s)
	}

	return &DockerPlugin{
		client:           client,
		prefix:           prefix,
		enableTotal:      enableTotal,
		enableFailing:    enableFailing,
		failingStatusMap: failingStatusMap,
	}, nil
}

func (m *DockerPlugin) listContainer() ([]docker.APIContainers, error) {
	containers, err := m.client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to execute listContainers API: %w", err)
	}

	return containers, nil
}

func (m *DockerPlugin) FetchMetrics() (map[string]float64, error) {
	containers, err := m.listContainer()
	if err != nil {
		return nil, fmt.Errorf("failed to get the list of containers: %w", err)
	}

	metrics := map[string]float64{}
	if m.enableTotal {
		metrics[MetricNameTotal] = 0
	}
	if m.enableFailing {
		metrics[MetricNameFailing] = 0
	}

	for _, c := range containers {
		if _, ok := statesMap[c.State]; ok {
			metrics[c.State] += 1
			if _, ok := m.failingStatusMap[c.State]; ok && m.enableFailing {
				metrics[MetricNameFailing] += 1
			}
			continue
		}

		if c.State != MetricNameRunning {
			return nil, fmt.Errorf("met unknown state of docker container. state: %s", c.State)
		}
		var metricName string
		status := getHealthCheckStatus(c.Status)
		switch status {
		case StatusStarting:
			metricName = MetricNameRunningStarting
		case StatusHealthy:
			metricName = MetricNameRunningHealthy
		case StatusUnhealthy:
			metricName = MetricNameRunningUnhealthy
		default:
			metricName = MetricNameRunning
		}
		metrics[metricName] += 1
		if _, ok := m.failingStatusMap[metricName]; ok && m.enableFailing {
			metrics[MetricNameFailing] += 1
		}
	}
	if m.enableTotal {
		metrics[MetricNameTotal] = float64(len(containers))
	}
	return metrics, nil
}

func (m *DockerPlugin) GraphDefinition() map[string]mp.Graphs {
	return map[string]mp.Graphs{
		"statuses": {
			Label: m.MetricKeyPrefix() + " Container Statuses",
			Unit:  mp.UnitInteger,
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
				{Name: "total", Label: "Total"},
				{Name: "failing", Label: "Failing"},
			},
		},
	}
}
func (m *DockerPlugin) MetricKeyPrefix() string {
	return m.prefix
}

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
