from collections.abc import Iterable
from typing import Any

from docker.client import DockerClient

from models.docker_resources import DockerResourceKind, ResourceSummary
from view_models.formatting import format_bytes, format_timestamp


class DockerService:
    def __init__(self, client: DockerClient):
        self.client = client

    def list_resources(self, kind: DockerResourceKind) -> list[ResourceSummary]:
        if kind == DockerResourceKind.CONTAINER:
            return self.list_containers()
        if kind == DockerResourceKind.IMAGE:
            return self.list_images()
        if kind == DockerResourceKind.VOLUME:
            return self.list_volumes()
        if kind == DockerResourceKind.NETWORK:
            return self.list_networks()
        raise ValueError(f"Unsupported resource kind: {kind}")

    def list_containers(self) -> list[ResourceSummary]:
        containers = self.client.containers.list(all=True)
        summaries = []
        for container in containers:
            attrs = getattr(container, "attrs", {})
            summaries.append(
                ResourceSummary(
                    kind=DockerResourceKind.CONTAINER,
                    id=container.id,
                    name=container.name,
                    raw=attrs,
                    columns={
                        "Image": _first_tag(getattr(container.image, "tags", [])),
                        "State": attrs.get("State", {}).get(
                            "Status", getattr(container, "status", "-")
                        ),
                        "Status": getattr(container, "status", "-"),
                        "Ports": _format_ports(
                            attrs.get("NetworkSettings", {}).get("Ports", {})
                        ),
                        "Created": format_timestamp(attrs.get("Created")),
                    },
                )
            )
        return summaries

    def list_images(self) -> list[ResourceSummary]:
        images = self.client.images.list()
        summaries = []
        for image in images:
            attrs = getattr(image, "attrs", {})
            repository, tag = _split_image_tag(_first_tag(getattr(image, "tags", [])))
            summaries.append(
                ResourceSummary(
                    kind=DockerResourceKind.IMAGE,
                    id=image.id,
                    name=repository,
                    raw=attrs,
                    columns={
                        "Repository": repository,
                        "Tag": tag,
                        "Image ID": image.id.replace("sha256:", "")[:12],
                        "Size": format_bytes(attrs.get("Size")),
                        "Created": format_timestamp(attrs.get("Created")),
                    },
                )
            )
        return summaries

    def list_volumes(self) -> list[ResourceSummary]:
        volumes = self.client.volumes.list()
        return [
            ResourceSummary(
                kind=DockerResourceKind.VOLUME,
                id=volume.name,
                name=volume.name,
                raw=getattr(volume, "attrs", {}),
                columns={
                    "Driver": getattr(volume, "attrs", {}).get("Driver", "-"),
                    "Scope": getattr(volume, "attrs", {}).get("Scope", "-"),
                    "Mountpoint": getattr(volume, "attrs", {}).get("Mountpoint", "-"),
                    "Created": format_timestamp(
                        getattr(volume, "attrs", {}).get("CreatedAt")
                    ),
                },
            )
            for volume in volumes
        ]

    def list_networks(self) -> list[ResourceSummary]:
        networks = self.client.networks.list()
        return [
            ResourceSummary(
                kind=DockerResourceKind.NETWORK,
                id=network.id,
                name=network.name,
                raw=getattr(network, "attrs", {}),
                columns={
                    "Driver": getattr(network, "attrs", {}).get("Driver", "-"),
                    "Scope": getattr(network, "attrs", {}).get("Scope", "-"),
                    "Flags": _network_flags(getattr(network, "attrs", {})),
                    "Containers": str(
                        len(getattr(network, "attrs", {}).get("Containers") or {})
                    ),
                },
            )
            for network in networks
        ]

    def prune(self, kind: DockerResourceKind) -> dict[str, Any]:
        if kind == DockerResourceKind.CONTAINER:
            return self.client.containers.prune()
        if kind == DockerResourceKind.IMAGE:
            return self.client.images.prune()
        if kind == DockerResourceKind.VOLUME:
            return self.client.volumes.prune()
        if kind == DockerResourceKind.NETWORK:
            return self.client.networks.prune()
        raise ValueError(f"Unsupported prune kind: {kind}")


def _first_tag(tags: Iterable[str]) -> str:
    return next(iter(tags), "<none>:<none>")


def _split_image_tag(value: str) -> tuple[str, str]:
    if ":" not in value:
        return value, "-"
    repository, tag = value.rsplit(":", 1)
    return repository, tag


def _format_ports(ports: dict[str, Any]) -> str:
    formatted = []
    for container_port, bindings in ports.items():
        port = container_port.split("/", 1)[0]
        if not bindings:
            formatted.append(port)
            continue
        for binding in bindings:
            formatted.append(
                f"{binding.get('HostPort', '-')}-{port}".replace("-", "->", 1)
            )
    return ", ".join(formatted) if formatted else "-"


def _network_flags(attrs: dict[str, Any]) -> str:
    flags = []
    if attrs.get("Internal"):
        flags.append("internal")
    if attrs.get("Attachable"):
        flags.append("attachable")
    return ", ".join(flags) if flags else "-"
