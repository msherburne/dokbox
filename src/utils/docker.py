from docker.client import DockerClient
import docker
import os


def detect_docker_host() -> str:
    docker_host = os.environ.get("DOCKER_HOST")
    if docker_host:
        return docker_host

    client = docker.from_env()
    adapter = getattr(client.api, "_custom_adapter", None)
    socket_path = getattr(adapter, "socket_path", None)

    if socket_path:
        return f"unix://{socket_path}"

    return client.api.base_url


def load_docker_client(host: str) -> DockerClient:
    return docker.DockerClient(base_url=host)
