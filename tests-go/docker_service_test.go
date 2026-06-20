package testsgo

import (
	"context"
	"errors"
	"testing"

	"github.com/msherburne/dokbox/internal/docker"
)

type pingClientStub struct {
	err       error
	pingCalls int
}

func (c *pingClientStub) Ping(context.Context) error {
	c.pingCalls++
	return c.err
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
