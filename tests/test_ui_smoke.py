import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "src"))

from models.docker_resources import DockerResourceKind, ResourceSummary
from ui.widgets import ResourceTable, ShortcutBar, build_resource_table
from view_models.resources import ShortcutHint


def test_shortcut_bar_renders_contextual_hints():
    bar = ShortcutBar([ShortcutHint("Enter", "Details"), ShortcutHint("q", "Quit")])

    assert bar.renderable == "Enter Details  q Quit"


def test_build_resource_table_stores_columns_and_rows_for_mount():
    resources = [
        ResourceSummary(
            kind=DockerResourceKind.CONTAINER,
            id="abcdef1234567890",
            name="api",
            raw={},
            columns={
                "Image": "dokbox-api",
                "State": "running",
                "Status": "Up",
                "Ports": "8080->80",
                "Created": "-",
            },
        )
    ]

    table = build_resource_table(DockerResourceKind.CONTAINER, resources)

    assert isinstance(table, ResourceTable)
    assert table.cursor_type == "row"
    assert table.kind == DockerResourceKind.CONTAINER
    assert table.resources == resources


def test_app_receives_docker_service():
    from main import Dokbox
    from models.config import DokboxConfig

    service = object()
    app = Dokbox(
        DokboxConfig(docker_host="unix:///tmp/docker.sock"),
        docker_service=service,
    )

    assert app.docker_service is service


def test_container_detail_screen_stores_container_id():
    from ui.screens import ContainerDetailScreen

    service = object()
    screen = ContainerDetailScreen(service, "abc")

    assert screen.container_id == "abc"
    assert screen.docker_service is service


def test_container_tool_panes_store_container_context():
    from ui.screens import FilesPane, LogsPane, ShellPane

    service = object()

    assert LogsPane(service, "abc").container_id == "abc"
    assert ShellPane(service, "abc").container_id == "abc"
    assert FilesPane(service, "abc", "/").path == "/"
