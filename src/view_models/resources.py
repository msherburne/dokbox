from dataclasses import dataclass

from rich.text import Text

from models.docker_resources import DockerResourceKind, MetricSample, ResourceSummary
from view_models.formatting import format_ratio_bar


RESOURCE_COLUMNS = {
    DockerResourceKind.CONTAINER: [
        "Stack",
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

STACK_COLORS = (
    "cyan",
    "green",
    "yellow",
    "magenta",
    "bright_blue",
    "bright_green",
    "bright_cyan",
    "bright_magenta",
    "orange3",
    "turquoise2",
    "deep_sky_blue1",
    "spring_green3",
)


@dataclass(frozen=True)
class ShortcutHint:
    key: str
    label: str


@dataclass(frozen=True)
class PaletteEntry:
    kind: DockerResourceKind
    resource_id: str
    title: str
    subtitle: str
    search_text: str


@dataclass(frozen=True)
class MetricCard:
    title: str
    value: str
    bar: str


def get_resource_columns(kind: DockerResourceKind) -> list[str]:
    return list(RESOURCE_COLUMNS[kind])


def resource_to_row(summary: ResourceSummary) -> list[str]:
    if summary.kind == DockerResourceKind.CONTAINER:
        return [
            stack_label(summary),
            container_cell(summary, summary.name),
            container_cell(summary, summary.columns.get("Image", "-")),
            container_cell(summary, summary.columns.get("State", "-")),
            container_cell(summary, summary.columns.get("Status", "-")),
            container_cell(summary, summary.columns.get("Ports", "-")),
            container_cell(summary, summary.columns.get("Created", "-")),
        ]
    return [
        summary.name if column == "Name" else summary.columns.get(column, "-")
        for column in get_resource_columns(summary.kind)
    ]


def stack_label(summary: ResourceSummary) -> Text:
    group_name = summary.group or "Ungrouped"
    if not summary.group:
        return Text(group_name, style="dim")
    color = STACK_COLORS[sum(ord(char) for char in group_name) % len(STACK_COLORS)]
    return Text(group_name, style=f"bold {color}")


def container_cell(summary: ResourceSummary, value: str) -> Text:
    if _is_down_container(summary):
        return Text(value, style="dim")
    return Text(value)


def _is_down_container(summary: ResourceSummary) -> bool:
    return summary.columns.get("State", "").casefold() != "running"


def get_shortcuts(context: str) -> list[ShortcutHint]:
    shortcuts = {
        "resource-tabs": [
            ShortcutHint("Left/Right", "Switch Tabs"),
            ShortcutHint("Enter", "Focus Table"),
            ShortcutHint("q", "Quit"),
        ],
        "containers-table": [
            ShortcutHint("Up/Down", "Rows"),
            ShortcutHint("Enter", "Open Details"),
            ShortcutHint("o", "Actions"),
            ShortcutHint("p", "Prune"),
            ShortcutHint("q", "Back To Tabs"),
        ],
        "images-table": [
            ShortcutHint("Up/Down", "Rows"),
            ShortcutHint("Enter", "Open Details"),
            ShortcutHint("p", "Prune"),
            ShortcutHint("q", "Back To Tabs"),
        ],
        "volumes-table": [
            ShortcutHint("Up/Down", "Rows"),
            ShortcutHint("Enter", "Open Details"),
            ShortcutHint("p", "Prune"),
            ShortcutHint("q", "Back To Tabs"),
        ],
        "networks-table": [
            ShortcutHint("Up/Down", "Rows"),
            ShortcutHint("Enter", "Open Details"),
            ShortcutHint("p", "Prune"),
            ShortcutHint("q", "Back To Tabs"),
        ],
        "containers": [
            ShortcutHint("Enter", "Details"),
            ShortcutHint("s", "Start/Stop"),
            ShortcutHint("r", "Restart"),
            ShortcutHint("x", "Remove"),
            ShortcutHint("p", "Prune"),
            ShortcutHint("/", "Filter"),
            ShortcutHint("q", "Quit"),
        ],
        "images": [
            ShortcutHint("Enter", "Details"),
            ShortcutHint("x", "Remove"),
            ShortcutHint("p", "Prune"),
            ShortcutHint("/", "Filter"),
            ShortcutHint("q", "Quit"),
        ],
        "volumes": [
            ShortcutHint("Enter", "Details"),
            ShortcutHint("x", "Remove"),
            ShortcutHint("p", "Prune"),
            ShortcutHint("/", "Filter"),
            ShortcutHint("q", "Quit"),
        ],
        "networks": [
            ShortcutHint("Enter", "Details"),
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
        "container-details": [
            ShortcutHint("Left/Right", "Switch Pane"),
            ShortcutHint("q", "Back"),
        ],
        "command-palette": [
            ShortcutHint("Up/Down", "Select"),
            ShortcutHint("Enter", "Open"),
            ShortcutHint("Esc", "Close"),
        ],
    }
    return shortcuts[context]


def build_metric_card(metric: MetricSample) -> MetricCard:
    return MetricCard(
        title=metric.name, value=metric.label, bar=format_ratio_bar(metric.ratio)
    )


def build_container_palette_entries(
    resources: list[ResourceSummary],
) -> list[PaletteEntry]:
    entries: list[PaletteEntry] = []
    for resource in resources:
        stack = resource.group or "Ungrouped"
        image = resource.columns.get("Image", "-")
        status = resource.columns.get("Status", resource.columns.get("State", "-"))
        state = resource.columns.get("State", "")
        entries.append(
            PaletteEntry(
                kind=resource.kind,
                resource_id=resource.id,
                title=resource.name,
                subtitle=f"{stack} | {image} | {status}",
                search_text=" ".join(
                    part.casefold()
                    for part in [resource.name, stack, image, state, status]
                    if part
                ),
            )
        )
    return entries


def filter_palette_entries(
    entries: list[PaletteEntry], query: str
) -> list[PaletteEntry]:
    tokens = [token for token in query.casefold().split() if token]
    if not tokens:
        return entries
    return [
        entry
        for entry in entries
        if all(token in entry.search_text for token in tokens)
    ]
