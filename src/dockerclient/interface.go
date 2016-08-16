package dockerclient

import (
	"github.com/Dataman-Cloud/rolex/src/dockerclient/model"
	opt "github.com/Dataman-Cloud/rolex/src/model"

	docker "github.com/Dataman-Cloud/go-dockerclient"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/swarm"
	"golang.org/x/net/context"
)

type DockerClientInterface interface {
	Ping() error
	InspectSwarm() (swarm.Swarm, error)
	ManagerInfo() (types.Info, error)

	NodeList(opts types.NodeListOptions) ([]swarm.Node, error)
	NodeInspect(nodeId string) (swarm.Node, error)
	NodeRemove(nodeId string) error
	UpdateNode(nodeId string, opts opt.UpdateOptions) error

	InspectContainer(id string) (*docker.Container, error)
	ListContainers(opts docker.ListContainersOptions) ([]docker.APIContainers, error)
	RemoveContainer(opts docker.RemoveContainerOptions) error
	LogsContainer(nodeId, containerId string, message chan string)
	StatsContainer(nodeId, containerId string, stats chan *model.ContainerStat)

	ConnectNetwork(id string, opts docker.NetworkConnectionOptions) error
	CreateNetwork(opts docker.CreateNetworkOptions) (*docker.Network, error)
	DisconnectNetwork(id string, opts docker.NetworkConnectionOptions) error
	InspectNetwork(id string) (*docker.Network, error)
	ListNetworks(opts docker.NetworkFilterOpts) ([]docker.Network, error)
	RemoveNetwork(id string) error

	CreateNodeNetwork(ctx context.Context, opts docker.CreateNetworkOptions) (*docker.Network, error)
	ConnectNodeNetwork(ctx context.Context, networkID string, opts docker.NetworkConnectionOptions) error
	DisconnectNodeNetwork(ctx context.Context, networkID string, opts docker.NetworkConnectionOptions) error
	InspectNodeNetwork(ctx context.Context, networkID string) (*docker.Network, error)
	ListNodeNetworks(ctx context.Context, opts docker.NetworkFilterOpts) ([]docker.Network, error)

	InspectVolume(nodeId, name string) (*docker.Volume, error)
	ListVolumes(nodeId string, opts docker.ListVolumesOptions) ([]docker.Volume, error)
	CreateVolume(nodeId string, opts docker.CreateVolumeOptions) (*docker.Volume, error)
	RemoveVolume(nodeId string, name string) error

	ListImages(nodeId string, opts docker.ListImagesOptions) ([]docker.APIImages, error)
	InspectImage(nodeId, imageId string) (*docker.Image, error)
	ImageHistory(nodeId, imageId string) ([]docker.ImageHistory, error)

	DeployStack(bundle *model.Bundle) error
	ListStack() ([]Stack, error)
	ListStackService(namespace string, opts types.ServiceListOptions) ([]ServiceStatus, error)
	InspectStack(namespace string) (*model.Bundle, error)
	RemoveStack(namespace string) error
	FilterServiceByStack(namespace string, opts types.ServiceListOptions) ([]swarm.Service, error)
	ConvertStackService(swarmService swarm.ServiceSpec) model.RolexService
	GetStackGroup(namespace string) (uint64, error)

	CreateService(service swarm.ServiceSpec, options types.ServiceCreateOptions) (types.ServiceCreateResponse, error)
	ListServiceSpec(options types.ServiceListOptions) ([]swarm.Service, error)
	ListService(options types.ServiceListOptions) ([]ServiceStatus, error)
	GetServicesStatus(services []swarm.Service) ([]ServiceStatus, error)
	RemoveService(serviceID string) error
	UpdateService(serviceID string, version swarm.Version, service swarm.ServiceSpec, header map[string][]string) error
	ScaleService(serviceID string, serviceScale ServiceScale) error
	InspectServiceWithRaw(serviceID string) (swarm.Service, error)
	ServiceAddLabel(serviceID string, labels map[string]string) error
	ServiceRemoveLabel(serviceID string, labels []string) error
}
