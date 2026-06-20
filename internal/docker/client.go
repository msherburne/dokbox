package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
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

func (c *Client) ContainerMetrics(containerID string) ([]domain.MetricSample, error) {
	statsReader, err := c.api.ContainerStatsOneShot(context.Background(), containerID)
	if err != nil {
		return nil, err
	}
	defer statsReader.Body.Close()

	var stats container.StatsResponse
	if err := json.NewDecoder(statsReader.Body).Decode(&stats); err != nil {
		return nil, err
	}

	inspect, err := c.api.ContainerInspect(context.Background(), containerID)
	if err != nil {
		return nil, err
	}

	cpuValue := calculateCPUCores(stats)
	cpuLimit := cpuLimitCores(inspect)
	memoryUsage := float64(stats.MemoryStats.Usage)
	memoryLimit := memoryLimit(inspect, stats)
	networkRX := networkTotal(stats, "rx")
	diskRead := blkioTotal(stats, "Read")

	metrics := []domain.MetricSample{
		{Name: "CPU", Value: cpuValue, Limit: cpuLimit, Unit: "cores", Label: cpuLabel(cpuValue, cpuLimit)},
		{Name: "Memory", Value: memoryUsage, Limit: memoryLimit, Unit: "bytes", Label: fmt.Sprintf("%s / %s", formatBytes(memoryUsage), formatBytes(valueOrDefault(memoryLimit)))},
		{Name: "Network RX", Value: networkRX, Unit: "bytes", Label: formatBytes(networkRX)},
		{Name: "Disk Read", Value: diskRead, Unit: "bytes", Label: formatBytes(diskRead)},
	}

	return metrics, nil
}

func calculateCPUCores(stats container.StatsResponse) float64 {
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)
	onlineCPUs := float64(stats.CPUStats.OnlineCPUs)
	if onlineCPUs == 0 {
		onlineCPUs = float64(len(stats.CPUStats.CPUUsage.PercpuUsage))
	}
	if onlineCPUs == 0 || systemDelta <= 0 {
		return 0
	}
	return (cpuDelta / systemDelta) * onlineCPUs
}

func cpuLimitCores(inspect container.InspectResponse) *float64 {
	if inspect.HostConfig == nil {
		return nil
	}
	if inspect.HostConfig.NanoCPUs > 0 {
		limit := float64(inspect.HostConfig.NanoCPUs) / 1_000_000_000
		return &limit
	}
	if inspect.HostConfig.CPUQuota > 0 && inspect.HostConfig.CPUPeriod > 0 {
		limit := float64(inspect.HostConfig.CPUQuota) / float64(inspect.HostConfig.CPUPeriod)
		return &limit
	}
	if inspect.HostConfig.CpusetCpus != "" {
		limit := float64(len(expandCPUSet(inspect.HostConfig.CpusetCpus)))
		return &limit
	}
	return nil
}

func memoryLimit(inspect container.InspectResponse, stats container.StatsResponse) *float64 {
	if inspect.HostConfig != nil && inspect.HostConfig.Memory > 0 {
		limit := float64(inspect.HostConfig.Memory)
		return &limit
	}
	if stats.MemoryStats.Limit > 0 {
		limit := float64(stats.MemoryStats.Limit)
		return &limit
	}
	return nil
}

func cpuLabel(value float64, limit *float64) string {
	if limit != nil && *limit > 0 {
		return fmt.Sprintf("%.0f%% of %g cores", (value / *limit) * 100, *limit)
	}
	return fmt.Sprintf("%.2f host cores", value)
}

func networkTotal(stats container.StatsResponse, direction string) float64 {
	var total uint64
	for _, network := range stats.Networks {
		if direction == "rx" {
			total += network.RxBytes
		} else {
			total += network.TxBytes
		}
	}
	return float64(total)
}

func blkioTotal(stats container.StatsResponse, op string) float64 {
	var total uint64
	for _, entry := range stats.BlkioStats.IoServiceBytesRecursive {
		if entry.Op == op {
			total += entry.Value
		}
	}
	return float64(total)
}

func expandCPUSet(value string) []int {
	var cpus []int
	for _, part := range strings.Split(value, ",") {
		if strings.Contains(part, "-") {
			bounds := strings.SplitN(part, "-", 2)
			start, _ := strconv.Atoi(bounds[0])
			end, _ := strconv.Atoi(bounds[1])
			for current := start; current <= end; current++ {
				cpus = append(cpus, current)
			}
			continue
		}
		if part == "" {
			continue
		}
		cpu, _ := strconv.Atoi(part)
		cpus = append(cpus, cpu)
	}
	return cpus
}

func formatBytes(value float64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	index := 0
	for value >= 1024 && index < len(units)-1 {
		value /= 1024
		index++
	}
	if index == 0 {
		return fmt.Sprintf("%.0f %s", value, units[index])
	}
	return fmt.Sprintf("%.1f %s", math.Round(value*10)/10, units[index])
}

func valueOrDefault(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
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
