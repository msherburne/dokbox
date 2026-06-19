from textual.widgets import DataTable, Static

from models.docker_resources import DockerResourceKind, ResourceSummary
from view_models.resources import ShortcutHint, get_resource_columns, resource_to_row


class ShortcutBar(Static):
    def __init__(self, shortcuts: list[ShortcutHint]):
        self.shortcuts = shortcuts
        super().__init__(self._render_shortcuts())

    @property
    def renderable(self) -> str:
        return self._render_shortcuts()

    def _render_shortcuts(self) -> str:
        return "  ".join(
            f"{shortcut.key} {shortcut.label}" for shortcut in self.shortcuts
        )


class ResourceTable(DataTable):
    def __init__(self, kind: DockerResourceKind, resources: list[ResourceSummary]):
        self.kind = kind
        self.resources = resources
        super().__init__(cursor_type="row")

    def on_mount(self) -> None:
        self.add_columns(*get_resource_columns(self.kind))
        for resource in self.resources:
            self.add_row(*resource_to_row(resource), key=resource.id)


def build_resource_table(
    kind: DockerResourceKind, resources: list[ResourceSummary]
) -> ResourceTable:
    return ResourceTable(kind, resources)
