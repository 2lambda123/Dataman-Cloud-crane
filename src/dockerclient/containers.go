package dockerclient

import (
	"bufio"
	"io"

	"github.com/Dataman-Cloud/rolex/src/dockerclient/model"
	"github.com/Dataman-Cloud/rolex/src/util/rolexerror"

	log "github.com/Sirupsen/logrus"
	goclient "github.com/fsouza/go-dockerclient"
	"golang.org/x/net/context"
)

func (client *RolexDockerClient) ListContainers(ctx context.Context, opts goclient.ListContainersOptions) ([]goclient.APIContainers, error) {
	swarmNode, err := client.SwarmNode(ctx)
	if err != nil {
		return nil, err
	}
	return swarmNode.ListContainers(opts)
}

func (client *RolexDockerClient) InspectContainer(ctx context.Context, id string) (*goclient.Container, error) {
	swarmNode, err := client.SwarmNode(ctx)
	if err != nil {
		return nil, err
	}

	container, err := swarmNode.InspectContainer(id)
	if err != nil {
		err = SortingError(err)
	}

	return container, err
}

func (client *RolexDockerClient) RemoveContainer(ctx context.Context, opts goclient.RemoveContainerOptions) error {
	swarmNode, err := client.SwarmNode(ctx)
	if err != nil {
		return err
	}

	err = swarmNode.RemoveContainer(opts)
	if err != nil {
		err = SortingError(err)
	}

	return err
}

func (client *RolexDockerClient) KillContainer(ctx context.Context, opts goclient.KillContainerOptions) error {
	swarmNode, err := client.SwarmNode(ctx)
	if err != nil {
		return err
	}

	err = swarmNode.KillContainer(opts)
	if err != nil {
		err = SortingError(err)
	}

	return err
}

func (client *RolexDockerClient) RenameContainer(ctx context.Context, opts goclient.RenameContainerOptions) error {
	swarmNode, err := client.SwarmNode(ctx)
	if err != nil {
		return err
	}

	err = swarmNode.RenameContainer(opts)
	if err != nil {
		err = SortingError(err)
	}

	return err
}

func (client *RolexDockerClient) DiffContainer(ctx context.Context, containerID string) ([]goclient.Change, error) {
	swarmNode, err := client.SwarmNode(ctx)
	if err != nil {
		return nil, err
	}

	changes, err := swarmNode.ContainerChanges(containerID)
	if err != nil {
		err = SortingError(err)
	}

	return changes, err
}

func (client *RolexDockerClient) StopContainer(ctx context.Context, containerId string, timeout uint) error {
	swarmNode, err := client.SwarmNode(ctx)
	if err != nil {
		return err
	}

	err = swarmNode.StopContainer(containerId, timeout)
	if err != nil {
		err = SortingError(err)
	}

	return err
}

func (client *RolexDockerClient) StartContainer(ctx context.Context, containerID string, hostconfig *goclient.HostConfig) error {
	swarmNode, err := client.SwarmNode(ctx)
	if err != nil {
		return err
	}

	err = swarmNode.StartContainer(containerID, hostconfig)
	if err != nil {
		err = SortingError(err)
	}

	return err
}

func (client *RolexDockerClient) RestartContainer(ctx context.Context, containerId string, timeout uint) error {
	swarmNode, err := client.SwarmNode(ctx)
	if err != nil {
		return err
	}

	err = swarmNode.RestartContainer(containerId, timeout)
	if err != nil {
		err = SortingError(err)
	}

	return err
}

func (client *RolexDockerClient) PauseContainer(ctx context.Context, containerID string) error {
	swarmNode, err := client.SwarmNode(ctx)
	if err != nil {
		return err
	}

	err = swarmNode.PauseContainer(containerID)
	if err != nil {
		err = SortingError(err)
	}

	return err
}

func (client *RolexDockerClient) UnpauseContainer(ctx context.Context, containerID string) error {
	swarNode, err := client.SwarmNode(ctx)
	if err != nil {
		return err
	}

	err = swarNode.UnpauseContainer(containerID)
	if err != nil {
		err = SortingError(err)
	}

	return err
}

func (client *RolexDockerClient) ResizeContainerTTY(ctx context.Context, containerID string, height, width int) error {
	swarNode, err := client.SwarmNode(ctx)
	if err != nil {
		return err
	}
	return swarNode.ResizeContainerTTY(containerID, height, width)
}

func (client *RolexDockerClient) LogsContainer(ctx context.Context, containerId string, message chan string) {
	swarNode, err := client.SwarmNode(ctx)
	if err != nil {
		log.Error("read container log error: ", err)
		return
	}
	outrd, outwr := io.Pipe()
	errrd, errwr := io.Pipe()

	go logReader(outrd, message)
	go logReader(errrd, message)

	opts := goclient.LogsOptions{
		Container:    containerId,
		OutputStream: outwr,
		ErrorStream:  errwr,
		Stdout:       true,
		Stderr:       true,
		Follow:       true,
		Tail:         "0",
	}
	err = swarNode.Logs(opts)
	log.Infof("read container log error: %v", err)
}

func logReader(input *io.PipeReader, message chan string) {
	buf := bufio.NewReader(input)

	defer func() {
		if err := recover(); err != nil {
			log.Errorf("send container log to channel error: %v", err)
		}
		input.Close()
		return
	}()

	for {
		select {
		default:
			if line, err := buf.ReadBytes('\n'); err != nil {
				log.Errorf("container log read buffer error: %v", err)
				return
			} else {
				message <- string(line)
			}
		}
	}
}

func (client *RolexDockerClient) StatsContainer(ctx context.Context, opts model.ContainerStatOptions) error {
	swarNode, err := client.SwarmNode(ctx)
	if err != nil {
		return err
	}

	cId := opts.ID
	container, err := swarNode.InspectContainer(cId)
	if err != nil {
		return err
	}

	chnError := make(chan error, 1)
	defer close(chnError)

	statOpts := goclient.StatsOptions{
		ID:     cId,
		Stats:  opts.Stats,
		Stream: opts.Stream,
		Done:   opts.Done,
	}
	go func() {
		chnError <- swarNode.Stats(statOpts)
	}()

	containerStat := &model.ContainerStat{
		NodeId:      container.Config.Labels["com.docker.swarm.node.id"],
		ServiceId:   container.Config.Labels["com.docker.swarm.service.id"],
		ServiceName: container.Config.Labels["com.docker.swarm.service.name"],
		TaskId:      container.Config.Labels["com.docker.swarm.task.id"],
		TaskName:    container.Config.Labels["com.docker.swarm.task.name"],
		ContainerId: container.ID,
	}

	for {
		select {
		case streamErr := <-chnError:
			return &rolexerror.ContainerStatsStopError{ID: cId, Err: streamErr}
		case stat := <-opts.Stats:
			containerStat.Stat = stat
			opts.RolexContainerStats <- containerStat
		}
	}
}
