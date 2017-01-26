package docker

import (
	"sort"
	"strconv"
	"time"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/signalfx/neo-agent/plugins"
	"github.com/signalfx/neo-agent/services"
	"golang.org/x/net/context"
)

const (
	defaultHostURL = "unix:///var/run/docker.sock"
	userAgent      = "signalfx-agent"
	version        = "v1.22"
)

// Docker observer plugin
type Docker struct {
	plugins.Plugin
}

// NewDocker constructor
func NewDocker(configuration map[string]string) *Docker {
	return &Docker{plugins.NewPlugin("docker", configuration)}
}

// Discover services from querying docker api
func (docker *Docker) Discover() (services.ServiceInstances, error) {

	defaultHeaders := map[string]string{"User-Agent": userAgent}
	hostURL := defaultHostURL
	if configVal, ok := docker.GetConfig("hosturl"); ok {
		hostURL = configVal
	}

	cli, err := client.NewClient(hostURL, version, nil, defaultHeaders)
	if err != nil {
		return nil, err
	}

	options := types.ContainerListOptions{All: true}
	containers, err := cli.ContainerList(context.Background(), options)
	if err != nil {
		return nil, err
	}

	instances := make(services.ServiceInstances, 0)

	for _, c := range containers {
		serviceContainer := services.NewServiceContainer(c.ID, c.Names, c.Image, "", c.Command, c.State, c.Labels)
		for _, port := range c.Ports {
			servicePort := services.NewServicePort(port.IP, port.Type, uint16(port.PrivatePort), uint16(port.PublicPort))
			id := docker.String() + c.ID + "-" + strconv.Itoa(port.PrivatePort)
			si := services.NewServiceInstance(id, nil, serviceContainer, nil, servicePort, time.Now())
			instances = append(instances, *si)
		}
	}

	sort.Sort(instances)

	return instances, nil
}
