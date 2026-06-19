from dataclasses import dataclass, field
from enum import Enum
from typing import Any, Iterator


class DockerResourceKind(str, Enum):
    CONTAINER = "container"
    IMAGE = "image"
    VOLUME = "volume"
    NETWORK = "network"


@dataclass(frozen=True)
class ResourceSummary:
    kind: DockerResourceKind
    id: str
    name: str
    raw: dict[str, Any]
    columns: dict[str, str] = field(default_factory=dict)
    group: str | None = None
    is_group: bool = False

    @property
    def short_id(self) -> str:
        return self.id[:12]


@dataclass(frozen=True)
class ResourceDetail:
    summary: ResourceSummary
    fields: dict[str, str]


@dataclass(frozen=True)
class MetricSample:
    name: str
    value: float
    limit: float | None
    unit: str
    label: str

    @property
    def ratio(self) -> float | None:
        if self.limit in (None, 0):
            return None
        return max(0.0, min(self.value / self.limit, 1.0))


@dataclass(frozen=True)
class FileEntry:
    path: str
    name: str
    is_dir: bool
    size: int
    mode: str


@dataclass(frozen=True)
class LogLine:
    text: str


@dataclass(frozen=True)
class ExecSession:
    command: tuple[str, ...]
    stream: Iterator[bytes]
