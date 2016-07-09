package dockerclient

import (
	goclient "github.com/fsouza/go-dockerclient"
)

func (client *RolexDockerClient) ListContainers(opts goclient.ListContainersOptions) ([]goclient.APIContainers, error) {
	return client.DockerClient.ListContainers(opts)
}

func (client *RolexDockerClient) InspectContainer(id string) (*goclient.Container, error) {
	return client.DockerClient.InspectContainer(id)
}

func (client *RolexDockerClient) RemoveContainer(opts goclient.RemoveContainerOptions) error {
	return client.DockerClient.RemoveContainer(opts)
}
