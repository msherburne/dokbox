package docker

import (
	"context"

	"github.com/msherburne/dokbox/internal/domain"
)

type PingClient interface {
	Ping(context.Context) error
}

type Service struct {
	client PingClient
}

func NewService(client PingClient) *Service {
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
