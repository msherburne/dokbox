package docker

import (
	"context"
	"os"
	"runtime"
	"time"

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
