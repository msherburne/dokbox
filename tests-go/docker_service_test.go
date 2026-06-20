package testsgo

import (
	"context"
	"errors"
	"testing"

	"github.com/msherburne/dokbox/internal/docker"
	"github.com/msherburne/dokbox/internal/domain"
)

type pingClientStub struct {
	err       error
	pingCalls int
	startedID string
	logLines  []domain.LogLine
}

func (c *pingClientStub) Ping(context.Context) error {
	c.pingCalls++
	return c.err
}

func (c *pingClientStub) StartContainer(containerID string) error {
	c.startedID = containerID
	return nil
}

func (c *pingClientStub) StopContainer(string) error {
	return nil
}

func (c *pingClientStub) RestartContainer(string) error {
	return nil
}

func (c *pingClientStub) RemoveResource(domain.ResourceKind, string) error {
	return nil
}

func (c *pingClientStub) Prune(domain.ResourceKind) error {
	return nil
}

func (c *pingClientStub) ContainerLogs(string, int) ([]domain.LogLine, error) {
	return c.logLines, nil
}

func TestCheckConnectionReturnsConnectedAfterPing(t *testing.T) {
	client := &pingClientStub{}
	service := docker.NewService(client)

	result := service.CheckConnection(context.Background())

	if client.pingCalls != 1 {
		t.Fatalf("expected ping once, got %d", client.pingCalls)
	}

	if !result.OK {
		t.Fatalf("expected OK status, got %#v", result)
	}

	if result.Message != "Connected" {
		t.Fatalf("expected connected message, got %q", result.Message)
	}
}

func TestCheckConnectionReturnsErrorMessageOnPingFailure(t *testing.T) {
	client := &pingClientStub{err: errors.New("cannot connect")}
	service := docker.NewService(client)

	result := service.CheckConnection(context.Background())

	if result.OK {
		t.Fatalf("expected failed status, got %#v", result)
	}

	if result.Message != "cannot connect" {
		t.Fatalf("expected error message, got %q", result.Message)
	}
}

func TestStatusCheckerReturnsFailureWhenClientBootstrapFails(t *testing.T) {
	checker := docker.NewStatusChecker("://not-a-valid-docker-host")

	result := checker.CheckConnection(context.Background())

	if result.OK {
		t.Fatalf("expected failed bootstrap status, got %#v", result)
	}

	if result.Message == "" {
		t.Fatalf("expected bootstrap failure message, got %#v", result)
	}
}

func TestStartContainerDelegatesToClient(t *testing.T) {
	client := &pingClientStub{}
	service := docker.NewService(client)

	if err := service.StartContainer("container-1"); err != nil {
		t.Fatalf("expected start to succeed, got %v", err)
	}

	if client.startedID != "container-1" {
		t.Fatalf("expected container start delegation, got %q", client.startedID)
	}
}

func TestContainerLogsDelegatesToClient(t *testing.T) {
	client := &pingClientStub{
		logLines: []domain.LogLine{{Text: "booting"}, {Text: "ready"}},
	}
	service := docker.NewService(client)

	lines, err := service.ContainerLogs("container-1", 200)
	if err != nil {
		t.Fatalf("expected logs to succeed, got %v", err)
	}

	if len(lines) != 2 || lines[0].Text != "booting" || lines[1].Text != "ready" {
		t.Fatalf("unexpected log lines: %#v", lines)
	}
}
