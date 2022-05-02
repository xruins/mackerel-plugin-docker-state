package docker

// Status represents healthcheck status of `running` containers
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

	MetricNameRunning          = "running"
	MetricNameRunningStarting  = "running_starting"
	MetricNameRunningHealthy   = "running_healthy"
	MetricNameRunningUnhealthy = "running_unhealthy"
	MetricNameTotal            = "total"
	MetricNameFailing          = "failing"
)

// statusesWithoutRunning is possible statuses except `running` states
var statusesWithoutRunning = []string{
	"created",
	"restarting",
	"exited",
	"paused",
	"dead",
}

// allStatuses is all possible status of docker container
var allStatuses = append(
	statusesWithoutRunning,
	MetricNameRunningStarting,
	MetricNameRunningHealthy,
	MetricNameRunningUnhealthy,
)

// statesMap is the map of statusesWithoutRunning
var statesMap map[string]struct{}

func init() {
	statesMap = make(map[string]struct{}, len(statusesWithoutRunning))
	for _, s := range statusesWithoutRunning {
		statesMap[s] = struct{}{}
	}
}
