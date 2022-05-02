package main

import (
	"flag"
	"log"

	mp "github.com/mackerelio/go-mackerel-plugin"
	"github.com/xruins/mackerel-plugin-docker-state/lib/docker"
)

func main() {
	optHost := flag.String("host", "unix://var/run/docker.sock", "host to Docker API")
	optPrefix := flag.String("metric-key-prefix", "docker", "Metric key prefix")
	flag.Parse()

	d, err := docker.NewDockerPlugin(*optHost, *optPrefix)
	if err != nil {
		log.Fatalf("failed to initalize plugin. err: %s", err)
	}
	plugin := mp.NewMackerelPlugin(d)
	plugin.Run()
}
