from dataclasses import dataclass

from models.docker_resources import DockerResourceKind, MetricSample, ResourceSummary
from view_models.formatting import format_ratio_bar


RESOURCE_COLUMNS = {
    DockerResourceKind.CONTAINER: [
        "Name",
        "Image",
        "State",
        "Status",
        "Ports",
        "Created",
    ],
    DockerResourceKind.IMAGE: ["Repository", "Tag", "Image ID", "Size", "Created"],
    DockerResourceKind.VOLUME: ["Name", "Driver", "Scope", "Mountpoint", "Created"],
    DockerResourceKind.NETWORK: ["Name", "Driver", "Scope", "Flags", "Containers"],
}


@dataclass(frozen=True)
class ShortcutHint:
    key: str
    label: str


@dataclass(frozen=True)
class MetricCard:
    title: str
    value: str
    bar: str


def get_resource_columns(kind: DockerResourceKind) -> list[str]:
    return list(RESOURCE_COLUMNS[kind])


def resource_to_row(summary: ResourceSummary) -> list[str]:
    return [
        summary.name if column == "Name" else summary.columns.get(column, "-")
        for column in get_resource_columns(summary.kind)
    ]


def get_shortcuts(context: str) -> list[ShortcutHint]:
    shortcuts = {
        "containers": [
            ShortcutHint("Enter", "Details"),
            ShortcutHint("s", "Start/Stop"),
            ShortcutHint("r", "Restart"),
            ShortcutHint("x", "Remove"),
            ShortcutHint("p", "Prune"),
            ShortcutHint("/", "Filter"),
            ShortcutHint("q", "Quit"),
        ],
        "files": [
            ShortcutHint("Enter", "Open"),
            ShortcutHint("Backspace", "Up"),
            ShortcutHint("/", "Find"),
            ShortcutHint("q", "Back"),
        ],
    }
    return shortcuts[context]


def build_metric_card(metric: MetricSample) -> MetricCard:
    return MetricCard(
        title=metric.name, value=metric.label, bar=format_ratio_bar(metric.ratio)
    )
