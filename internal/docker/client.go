package docker

import (
	"os"
	"runtime"
)

const (
	defaultUnixDockerHost    = "unix:///var/run/docker.sock"
	defaultWindowsDockerHost = "npipe:////./pipe/docker_engine"
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
