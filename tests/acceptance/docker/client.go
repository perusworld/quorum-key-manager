package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/consensys/quorum-key-manager/pkg/errors"
	"github.com/consensys/quorum-key-manager/src/infra/log"
	"github.com/consensys/quorum-key-manager/tests/acceptance/docker/config"
	"github.com/consensys/quorum-key-manager/tests/acceptance/docker/container"
	"github.com/consensys/quorum-key-manager/tests/acceptance/docker/container/compose"
	"github.com/docker/docker/api/types"
	dockercontainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

type Client struct {
	cli *client.Client

	composition *config.Composition
	factory     container.DockerContainerFactory
	containers  map[string]dockercontainer.ContainerCreateCreatedBody
	networks    map[string]string
	logger      log.Logger
}

type dockerAuth struct {
	Auths map[string]dockerAuthItem `json:"auths"`
}

type dockerAuthItem struct {
	Auth string `json:"auth"`
}

var dockerRegistries = []string{"https://index.docker.io/v2/", "index.docker.io/v2/", "https://index.docker.io/v1/", "index.docker.io/v1/"}

func NewClient(composition *config.Composition, logger log.Logger) (*Client, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		cli:         cli,
		composition: composition,
		factory:     compose.New(logger),
		containers:  make(map[string]dockercontainer.ContainerCreateCreatedBody),
		networks:    make(map[string]string),
		logger:      logger,
	}, nil
}

func (c *Client) Up(ctx context.Context, name, networkName string) error {
	logger := c.logger.With("container", name)

	containerCfg, hostCfg, networkCfg, err := c.factory.GenerateContainerConfig(ctx, c.composition.Containers[name])
	if err != nil {
		return err
	}

	// Pull image
	err = c.pullImage(ctx, containerCfg.Image)
	if err != nil {
		return err
	}

	// Create Docker container
	containerBody, err := c.cli.ContainerCreate(ctx, containerCfg, hostCfg, networkCfg, name)
	if err != nil {
		return err
	}
	c.containers[name] = containerBody

	// Connect to network and assign the alias as the name of the container
	if networkID, ok := c.networks[networkName]; networkName != "" && ok {
		err = c.cli.NetworkConnect(ctx, networkID, containerBody.ID, &network.EndpointSettings{
			Aliases:   []string{name},
			NetworkID: networkID,
		})

		if err != nil {
			return err
		}

		logger.Info("container connected to network", "network_id", networkID)
	} else if networkName != "" {
		errMsg := fmt.Sprintf("container %v cannot connected to network", name)
		logger.Error(errMsg, "network_id", networkID)
		return errors.DependencyFailureError(errMsg)
	}

	// Start docker container
	if err := c.cli.ContainerStart(ctx, containerBody.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	logger.Info("started container", "id", containerBody.ID)

	return nil
}

func (c *Client) Start(ctx context.Context, name string) error {
	logger := c.logger.With("container", name)

	containerBody, err := c.getContainer(name)
	if err != nil {
		return nil
	}

	// Start docker container
	if err := c.cli.ContainerStart(ctx, containerBody.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	logger.Info("started container", "id", containerBody.ID)

	return nil
}

func (c *Client) Stop(ctx context.Context, name string) error {
	logger := c.logger.With("container", name)

	containerBody, err := c.getContainer(name)
	if err != nil {
		return nil
	}

	if err := c.cli.ContainerStop(ctx, containerBody.ID, nil); err != nil {
		return err
	}

	logger.With("id", containerBody.ID).Info("stopped container")

	return nil
}

func (c *Client) WaitTillIsReady(ctx context.Context, name string, timeout time.Duration) error {
	logger := c.logger.With("container", name)

	err := c.factory.WaitForService(ctx, c.composition.Containers[name], timeout)
	if err != nil {
		logger.WithError(err).Error("failed to wait for service", "service", name)
		return err
	}

	return nil
}

func (c *Client) StartServiceAndWait(ctx context.Context, name string, timeout time.Duration) error {
	err := c.Start(ctx, name)
	if err != nil {
		return err
	}

	err = c.WaitTillIsReady(ctx, name, timeout)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Down(ctx context.Context, name string) error {
	logger := c.logger.With("container", name)

	containerBody, err := c.getContainer(name)
	if err != nil {
		logger.WithError(err).Warn("container %s cannot be found", name)
		if err := c.cli.ContainerStop(ctx, name, nil); err != nil {
			return err
		}
	}

	if err := c.cli.ContainerStop(ctx, containerBody.ID, nil); err != nil {
		return err
	}

	logger.Info("stopped container", "id", containerBody.ID)

	if err := c.cli.ContainerRemove(ctx, containerBody.ID, types.ContainerRemoveOptions{RemoveVolumes: true}); err != nil {
		return err
	}

	logger.Info("removed container", "id", containerBody.ID)

	return nil
}

func (c *Client) CreateNetwork(ctx context.Context, name string) error {
	logger := c.logger.With("network_name", name)

	createResponse, err := c.cli.NetworkCreate(ctx, name, types.NetworkCreate{Driver: "bridge"})
	if err != nil {
		return err
	}

	logger.Info("created network", "id", createResponse.ID)
	c.networks[name] = createResponse.ID
	return nil
}

func (c *Client) RemoveNetwork(ctx context.Context, name string) error {
	logger := c.logger.With("network_name", name)

	if networkID, ok := c.networks[name]; !ok {
		err := c.cli.NetworkRemove(ctx, networkID)
		if err != nil {
			return err
		}

		logger.Info("removed network", "network_id", networkID)
	} else {
		err := c.cli.NetworkRemove(ctx, name)
		if err != nil {
			return err
		}

		logger.Info("removed network")
	}

	return nil
}

func (c *Client) getContainer(name string) (dockercontainer.ContainerCreateCreatedBody, error) {
	containerBody, ok := c.containers[name]
	if !ok {
		return dockercontainer.ContainerCreateCreatedBody{}, fmt.Errorf("no container found with name '%v'", name)
	}

	return containerBody, nil
}

func (c *Client) pullImage(ctx context.Context, imageName string) error {
	logger := c.logger.With("image_name", imageName)

	cfg := types.ImagePullOptions{}
	dockerAthCfg := &dockerAuth{}
	if err := json.Unmarshal([]byte(os.Getenv("DOCKER_AUTH_CONFIG")), dockerAthCfg); err == nil {
		for _, reg := range dockerRegistries {
			if crdt, ok := dockerAthCfg.Auths[reg]; ok {
				cfg.RegistryAuth = crdt.Auth
				logger.Info("docker registry credential", "auth", crdt.Auth, "reg", reg)
				break
			}
		}
	}

	// Pull image
	events, err := c.cli.ImagePull(ctx, imageName, cfg)

	if err != nil {
		return err
	}

	d := json.NewDecoder(events)

	type Event struct {
		Status         string `json:"status"`
		Error          string `json:"error"`
		Progress       string `json:"progress"`
		ProgressDetail struct {
			Current int `json:"current"`
			Total   int `json:"total"`
		} `json:"progressDetail"`
	}

	var event *Event
	for {
		if err := d.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}

			return err
		}
	}

	if event != nil {
		if strings.Contains(event.Status, fmt.Sprintf("Downloaded newer image for %s", imageName)) {
			logger.Info("downloaded new image")
		}

		if strings.Contains(event.Status, fmt.Sprintf("Image is up to date for %s", imageName)) {
			logger.Info("image up to date")
		}
	}

	return nil
}
