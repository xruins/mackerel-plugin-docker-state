package main

import (
	"flag"
	"log"
	"strings"

	mp "github.com/mackerelio/go-mackerel-plugin"
	"github.com/xruins/mackerel-plugin-docker-state/lib/docker"
)

func main() {
	optHost := flag.String("host", "unix://var/run/docker.sock", "host to Docker API")
	optPrefix := flag.String("metric-key-prefix", "docker", "Metric key prefix")
	optTotal := flag.Bool("enable-total", true, "Enable output of the metric shows total count of containers")
	optFailing := flag.Bool("enable-failing", true, "Enable output of the metric shows count of containers in failing state")
	optFailingStates := flag.String("failing-states", "dead,exited,paused,running_unhealthy", "Docker states treat as failing state (comma separated string)")
	flag.Parse()

	d, err := docker.NewDockerPlugin(
		*optHost,
		*optPrefix,
		*optTotal,
		*optFailing,
		strings.Split(*optFailingStates, ","),
	)
	if err != nil {
		log.Fatalf("failed to initalize plugin. err: %s", err)
	}
	plugin := mp.NewMackerelPlugin(d)
	plugin.Run()
}
