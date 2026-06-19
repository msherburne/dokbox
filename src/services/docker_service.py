from collections.abc import Iterable, Iterator
from dataclasses import dataclass
from stat import S_ISDIR
from typing import Any

from docker.client import DockerClient

from models.docker_resources import (
    DockerResourceKind,
    ExecSession,
    FileEntry,
    LogLine,
    MetricSample,
    ResourceSummary,
)
from view_models.formatting import format_bytes, format_timestamp


@dataclass(frozen=True)
class ConnectionStatus:
    ok: bool
    message: str


class DockerService:
    def __init__(self, client: DockerClient):
        self.client = client

    def check_connection(self) -> ConnectionStatus:
        try:
            self.client.ping()
        except Exception as exc:
            return ConnectionStatus(ok=False, message=str(exc))
        return ConnectionStatus(ok=True, message="Connected")

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
                        "Image": _container_image(attrs),
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

    def start_container(self, container_id: str) -> None:
        self.client.containers.get(container_id).start()

    def stop_container(self, container_id: str) -> None:
        self.client.containers.get(container_id).stop()

    def restart_container(self, container_id: str) -> None:
        self.client.containers.get(container_id).restart()

    def get_container_metrics(self, container_id: str) -> list[MetricSample]:
        container = self.client.containers.get(container_id)
        stats = container.stats(stream=False)
        attrs = getattr(container, "attrs", {})
        cpu_value = _calculate_cpu_cores(stats)
        cpu_limit = _cpu_limit_cores(attrs)
        memory_usage = stats.get("memory_stats", {}).get("usage", 0)
        memory_limit = _memory_limit(attrs, stats)
        return [
            MetricSample(
                name="CPU",
                value=cpu_value,
                limit=cpu_limit,
                unit="cores",
                label=_cpu_label(cpu_value, cpu_limit),
            ),
            MetricSample(
                name="Memory",
                value=memory_usage,
                limit=memory_limit,
                unit="bytes",
                label=f"{format_bytes(memory_usage)} / {format_bytes(memory_limit)}",
            ),
            MetricSample(
                name="Network RX",
                value=_network_total(stats, "rx_bytes"),
                limit=None,
                unit="bytes",
                label=format_bytes(_network_total(stats, "rx_bytes")),
            ),
            MetricSample(
                name="Disk Read",
                value=_blkio_total(stats, "Read"),
                limit=None,
                unit="bytes",
                label=format_bytes(_blkio_total(stats, "Read")),
            ),
        ]

    def stream_logs(
        self, container_id: str, follow: bool = True, tail: int = 200
    ) -> Iterator[LogLine]:
        container = self.client.containers.get(container_id)
        for chunk in container.logs(stream=True, follow=follow, tail=tail):
            yield LogLine(text=chunk.decode(errors="replace").rstrip("\n"))

    def open_shell(self, container_id: str) -> ExecSession:
        container = self.client.containers.get(container_id)
        command = ["/bin/bash"]
        exit_code, _output = container.exec_run(["test", "-x", "/bin/bash"])
        if exit_code != 0:
            command = ["/bin/sh"]
        exec_info = self.client.api.exec_create(
            container_id, cmd=command, stdin=True, tty=True
        )
        stream = self.client.api.exec_start(
            exec_info["Id"], stream=True, socket=False, tty=True
        )
        return ExecSession(command=tuple(command), stream=stream)

    def list_container_path(self, container_id: str, path: str) -> list[FileEntry]:
        container = self.client.containers.get(container_id)
        _stream, stat = container.get_archive(path)
        mode = stat.get("mode", 0)
        return [
            FileEntry(
                path=path,
                name=stat.get("name") or path.rsplit("/", 1)[-1] or "/",
                is_dir=S_ISDIR(mode),
                size=stat.get("size", 0),
                mode=oct(mode),
            )
        ]

    def remove_resource(self, kind: DockerResourceKind, resource_id: str) -> None:
        if kind == DockerResourceKind.CONTAINER:
            self.client.containers.get(resource_id).remove()
            return
        if kind == DockerResourceKind.IMAGE:
            self.client.images.remove(resource_id)
            return
        if kind == DockerResourceKind.VOLUME:
            self.client.volumes.get(resource_id).remove()
            return
        if kind == DockerResourceKind.NETWORK:
            self.client.networks.get(resource_id).remove()
            return
        raise ValueError(f"Unsupported remove kind: {kind}")


def _first_tag(tags: Iterable[str]) -> str:
    return next(iter(tags), "<none>:<none>")


def _split_image_tag(value: str) -> tuple[str, str]:
    if ":" not in value:
        return value, "-"
    repository, tag = value.rsplit(":", 1)
    return repository, tag


def _container_image(attrs: dict[str, Any]) -> str:
    return attrs.get("Config", {}).get("Image") or attrs.get("Image") or "-"


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


def _calculate_cpu_cores(stats: dict[str, Any]) -> float:
    cpu_stats = stats.get("cpu_stats", {})
    precpu_stats = stats.get("precpu_stats", {})
    cpu_usage = cpu_stats.get("cpu_usage", {})
    cpu_delta = cpu_usage.get("total_usage", 0) - precpu_stats.get("cpu_usage", {}).get(
        "total_usage", 0
    )
    system_delta = cpu_stats.get("system_cpu_usage", 0) - precpu_stats.get(
        "system_cpu_usage", 0
    )
    online_cpus = (
        cpu_stats.get("online_cpus") or len(cpu_usage.get("percpu_usage", [])) or 1
    )
    if system_delta <= 0:
        return 0.0
    return (cpu_delta / system_delta) * online_cpus


def _cpu_limit_cores(attrs: dict[str, Any]) -> float | None:
    host_config = attrs.get("HostConfig", {})
    nano_cpus = host_config.get("NanoCpus")
    if nano_cpus:
        return nano_cpus / 1_000_000_000
    quota = host_config.get("CpuQuota")
    period = host_config.get("CpuPeriod")
    if quota and period and quota > 0 and period > 0:
        return quota / period
    cpuset = host_config.get("CpusetCpus")
    if cpuset:
        return float(len(_expand_cpuset(cpuset)))
    return None


def _expand_cpuset(value: str) -> set[int]:
    cpus = set()
    for part in value.split(","):
        if "-" in part:
            start, end = part.split("-", 1)
            cpus.update(range(int(start), int(end) + 1))
        elif part:
            cpus.add(int(part))
    return cpus


def _memory_limit(attrs: dict[str, Any], stats: dict[str, Any]) -> int | None:
    host_limit = attrs.get("HostConfig", {}).get("Memory")
    if host_limit:
        return host_limit
    return stats.get("memory_stats", {}).get("limit")


def _cpu_label(value: float, limit: float | None) -> str:
    if limit:
        return f"{value / limit:.0%} of {limit:g} cores"
    return f"{value:.2f} host cores"


def _network_total(stats: dict[str, Any], field: str) -> int:
    return sum(network.get(field, 0) for network in stats.get("networks", {}).values())


def _blkio_total(stats: dict[str, Any], op: str) -> int:
    return sum(
        entry.get("value", 0)
        for entry in stats.get("blkio_stats", {}).get("io_service_bytes_recursive", [])
        if entry.get("op") == op
    )
