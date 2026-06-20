package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	dockerclient "github.com/docker/docker/client"
	"github.com/msherburne/dokbox/internal/domain"
)

const (
	defaultUnixDockerHost    = "unix:///var/run/docker.sock"
	defaultWindowsDockerHost = "npipe:////./pipe/docker_engine"
	clientTimeout            = 2 * time.Second
)

func DetectDockerHost() string {
	if dockerHost := os.Getenv("DOCKER_HOST"); dockerHost != "" {
		return dockerHost
	}

	return defaultDockerHost()
}

func defaultDockerHost() string {
	if runtime.GOOS == "windows" {
		return defaultWindowsDockerHost
	}

	return defaultUnixDockerHost
}

type Client struct {
	api *dockerclient.Client
}

func NewClient(host string) (*Client, error) {
	apiClient, err := dockerclient.NewClientWithOpts(
		dockerclient.FromEnv,
		dockerclient.WithHost(host),
		dockerclient.WithTimeout(clientTimeout),
		dockerclient.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, err
	}

	return &Client{api: apiClient}, nil
}

func (c *Client) Ping(ctx context.Context) error {
	_, err := c.api.Ping(ctx)
	return err
}

func (c *Client) StartContainer(containerID string) error {
	return c.api.ContainerStart(context.Background(), containerID, container.StartOptions{})
}

func (c *Client) StopContainer(containerID string) error {
	return c.api.ContainerStop(context.Background(), containerID, container.StopOptions{})
}

func (c *Client) RestartContainer(containerID string) error {
	return c.api.ContainerRestart(context.Background(), containerID, container.StopOptions{})
}

func (c *Client) RemoveResource(kind domain.ResourceKind, resourceID string) error {
	switch kind {
	case domain.ResourceKindContainer:
		return c.api.ContainerRemove(context.Background(), resourceID, container.RemoveOptions{})
	case domain.ResourceKindImage:
		_, err := c.api.ImageRemove(context.Background(), resourceID, image.RemoveOptions{})
		return err
	case domain.ResourceKindVolume:
		return c.api.VolumeRemove(context.Background(), resourceID, false)
	case domain.ResourceKindNetwork:
		return c.api.NetworkRemove(context.Background(), resourceID)
	default:
		return fmt.Errorf("unsupported remove kind: %s", kind)
	}
}

func (c *Client) Prune(kind domain.ResourceKind) error {
	switch kind {
	case domain.ResourceKindContainer:
		_, err := c.api.ContainersPrune(context.Background(), filters.Args{})
		return err
	case domain.ResourceKindImage:
		_, err := c.api.ImagesPrune(context.Background(), filters.Args{})
		return err
	case domain.ResourceKindVolume:
		_, err := c.api.VolumesPrune(context.Background(), filters.Args{})
		return err
	case domain.ResourceKindNetwork:
		_, err := c.api.NetworksPrune(context.Background(), filters.Args{})
		return err
	default:
		return fmt.Errorf("unsupported prune kind: %s", kind)
	}
}

func (c *Client) ContainerLogs(containerID string, tail int) ([]domain.LogLine, error) {
	reader, err := c.api.ContainerLogs(
		context.Background(),
		containerID,
		container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Tail:       strconv.Itoa(tail),
		},
	)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	payload, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	text := strings.TrimSpace(string(payload))
	if text == "" {
		return nil, nil
	}

	rawLines := strings.Split(text, "\n")
	lines := make([]domain.LogLine, 0, len(rawLines))
	for _, line := range rawLines {
		lines = append(lines, domain.LogLine{Text: strings.TrimRight(line, "\r")})
	}

	return lines, nil
}

func (c *Client) Close() error {
	return c.api.Close()
}

type StatusChecker struct {
	host string
}

func NewStatusChecker(host string) *StatusChecker {
	return &StatusChecker{host: host}
}

func (s *StatusChecker) CheckConnection(ctx context.Context) domain.ConnectionStatus {
	client, err := NewClient(s.host)
	if err != nil {
		return domain.ConnectionStatus{
			OK:      false,
			Message: err.Error(),
		}
	}
	defer client.Close()

	return NewService(client).CheckConnection(ctx)
}
