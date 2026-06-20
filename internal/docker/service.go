package docker

import (
	"context"

	"github.com/msherburne/dokbox/internal/domain"
)

type ActionClient interface {
	Ping(context.Context) error
	StartContainer(containerID string) error
	StopContainer(containerID string) error
	RestartContainer(containerID string) error
	RemoveResource(kind domain.ResourceKind, resourceID string) error
	Prune(kind domain.ResourceKind) error
	ContainerLogs(containerID string, tail int) ([]domain.LogLine, error)
	ContainerMetrics(containerID string) ([]domain.MetricSample, error)
}

type Service struct {
	client ActionClient
}

func NewService(client ActionClient) *Service {
	return &Service{client: client}
}

func (s *Service) CheckConnection(ctx context.Context) domain.ConnectionStatus {
	if err := s.client.Ping(ctx); err != nil {
		return domain.ConnectionStatus{
			OK:      false,
			Message: err.Error(),
		}
	}

	return domain.ConnectionStatus{
		OK:      true,
		Message: "Connected",
	}
}

func (s *Service) StartContainer(containerID string) error {
	return s.client.StartContainer(containerID)
}

func (s *Service) StopContainer(containerID string) error {
	return s.client.StopContainer(containerID)
}

func (s *Service) RestartContainer(containerID string) error {
	return s.client.RestartContainer(containerID)
}

func (s *Service) RemoveResource(kind domain.ResourceKind, resourceID string) error {
	return s.client.RemoveResource(kind, resourceID)
}

func (s *Service) Prune(kind domain.ResourceKind) error {
	return s.client.Prune(kind)
}

func (s *Service) ContainerLogs(containerID string, tail int) ([]domain.LogLine, error) {
	return s.client.ContainerLogs(containerID, tail)
}

func (s *Service) ContainerMetrics(containerID string) ([]domain.MetricSample, error) {
	return s.client.ContainerMetrics(containerID)
}
