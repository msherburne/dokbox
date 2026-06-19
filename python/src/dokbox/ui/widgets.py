from rich.text import Text
from textual.widgets import DataTable, Static

from dokbox.models.docker_resources import DockerResourceKind, ResourceSummary
from dokbox.view_models.resources import (
    ShortcutHint,
    get_resource_columns,
    resource_to_row,
)


class ShortcutBar(Static):
    DEFAULT_CSS = """
    ShortcutBar {
        height: 1;
        padding: 0 1;
        background: $panel-darken-2;
        color: $text;
    }
    """

    def __init__(self, shortcuts: list[ShortcutHint], **kwargs):
        self.shortcuts = shortcuts
        super().__init__(self._render_shortcuts(), **kwargs)

    def set_shortcuts(self, shortcuts: list[ShortcutHint]) -> None:
        self.shortcuts = shortcuts
        self.update(self._render_shortcuts())

    @property
    def renderable(self) -> Text:
        return self._render_shortcuts()

    def _render_shortcuts(self) -> Text:
        rendered = Text(no_wrap=True, overflow="ellipsis")
        for index, shortcut in enumerate(self.shortcuts):
            if index:
                rendered.append(" | ", style="dim")
            rendered.append(_display_key(shortcut.key), style="bold bright_yellow")
            rendered.append(" ")
            rendered.append(shortcut.label, style="cyan")
        return rendered


class ResourceTable(DataTable):
    def __init__(self, kind: DockerResourceKind, resources: list[ResourceSummary]):
        self.kind = kind
        self.resources = (
            sorted(resources, key=_container_sort_key)
            if kind == DockerResourceKind.CONTAINER
            else resources
        )
        self.display_rows = (
            _with_stack_spacers(self.resources)
            if kind == DockerResourceKind.CONTAINER
            else list(self.resources)
        )
        super().__init__(cursor_type="row")

    def on_mount(self) -> None:
        self._load_rows()

    def refresh_resources(self, resources: list[ResourceSummary]) -> None:
        self.resources = (
            sorted(resources, key=_container_sort_key)
            if self.kind == DockerResourceKind.CONTAINER
            else resources
        )
        self.display_rows = (
            _with_stack_spacers(self.resources)
            if self.kind == DockerResourceKind.CONTAINER
            else list(self.resources)
        )
        self.clear(columns=False)
        self._load_rows()

    def _load_rows(self) -> None:
        self.add_columns(*get_resource_columns(self.kind))
        for index, resource in enumerate(self.display_rows):
            if resource is None:
                self.add_row(
                    *[""] * len(get_resource_columns(self.kind)), key=f"spacer-{index}"
                )
                continue
            self.add_row(*resource_to_row(resource), key=resource.id)
        for index, resource in enumerate(self.display_rows):
            if resource is not None:
                self.cursor_coordinate = (index, 0)
                break

    @property
    def selected_resource(self) -> ResourceSummary | None:
        if not self.display_rows:
            return None
        row_index = self.cursor_row
        if row_index < 0 or row_index >= len(self.display_rows):
            return None
        return self.display_rows[row_index]


def build_resource_table(
    kind: DockerResourceKind, resources: list[ResourceSummary]
) -> ResourceTable:
    return ResourceTable(kind, resources)


def _container_sort_key(summary: ResourceSummary) -> tuple[int, str, str]:
    if summary.group:
        return (0, summary.group.casefold(), summary.name.casefold())
    return (1, "zzzzzz", summary.name.casefold())


def _with_stack_spacers(
    resources: list[ResourceSummary],
) -> list[ResourceSummary | None]:
    display_rows: list[ResourceSummary | None] = []
    previous_group: str | None | object = object()
    for resource in resources:
        current_group = resource.group or "Ungrouped"
        if display_rows and current_group != previous_group:
            display_rows.append(None)
        display_rows.append(resource)
        previous_group = current_group
    return display_rows


def _display_key(key: str) -> str:
    symbols = {
        "Left/Right": "←/→",
        "Up/Down": "↑/↓",
        "Enter": "↵",
        "Backspace": "⌫",
    }
    return symbols.get(key, key)
