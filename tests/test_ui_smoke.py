import asyncio
import sys
from pathlib import Path

from rich.text import Text

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "src"))

from textual.app import App
from textual.containers import Container
from textual.widgets import Footer
from textual.widgets import TabbedContent
from textual.widgets._tabbed_content import ContentTabs

from models.docker_resources import (
    DockerResourceKind,
    LogLine,
    MetricSample,
    ResourceSummary,
)
from ui.screens import ResourceBrowserScreen
from ui.widgets import ResourceTable, ShortcutBar, build_resource_table
from view_models.resources import ShortcutHint


class _OkStatus:
    ok = True
    message = "Connected"


class _BrowserServiceStub:
    def __init__(self):
        self.performed_actions = []

    def check_connection(self):
        return _OkStatus()

    def list_resources(self, kind: DockerResourceKind):
        return [
            ResourceSummary(
                kind=kind,
                id=f"{kind.value}-1",
                name=f"{kind.value}-row",
                raw={},
                columns={
                    "Image": "-",
                    "State": "-",
                    "Status": "-",
                    "Ports": "-",
                    "Created": "-",
                },
            )
        ]

    def get_container_metrics(self, container_id: str):
        return [
            MetricSample(
                name="CPU",
                value=0.42,
                limit=1.0,
                unit="cores",
                label="42% of 1.0 core",
            )
        ]

    def stream_logs(self, container_id: str, follow: bool = False):
        return iter([LogLine(text="hello")])

    def list_container_path(self, container_id: str, path: str):
        return []

    def start_container(self, container_id: str):
        self.performed_actions.append(("start", container_id))

    def stop_container(self, container_id: str):
        self.performed_actions.append(("stop", container_id))

    def restart_container(self, container_id: str):
        self.performed_actions.append(("restart", container_id))

    def remove_resource(self, kind: DockerResourceKind, resource_id: str):
        self.performed_actions.append(("remove", kind, resource_id))

    def prune(self, kind: DockerResourceKind):
        self.performed_actions.append(("prune", kind))


class _BrowserApp(App[None]):
    def __init__(self):
        super().__init__()
        self.docker_service = _BrowserServiceStub()
        self.screen_under_test = ResourceBrowserScreen(self.docker_service)

    def on_mount(self) -> None:
        self.push_screen(self.screen_under_test)


def test_shortcut_bar_renders_contextual_hints():
    bar = ShortcutBar([ShortcutHint("Enter", "Details"), ShortcutHint("q", "Quit")])

    assert isinstance(bar.renderable, Text)
    assert bar.renderable.plain == "↵ Details | q Quit"


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


def test_build_resource_table_sorts_containers_by_stack_then_name():
    resources = [
        ResourceSummary(
            kind=DockerResourceKind.CONTAINER,
            id="3",
            name="zeta",
            raw={},
            columns={},
            group=None,
        ),
        ResourceSummary(
            kind=DockerResourceKind.CONTAINER,
            id="2",
            name="worker",
            raw={},
            columns={},
            group="odysseus",
        ),
        ResourceSummary(
            kind=DockerResourceKind.CONTAINER,
            id="1",
            name="api",
            raw={},
            columns={},
            group="atlas",
        ),
    ]

    table = build_resource_table(DockerResourceKind.CONTAINER, resources)

    assert [resource.name for resource in table.resources] == ["api", "worker", "zeta"]


def test_build_resource_table_inserts_spacer_rows_between_stacks():
    resources = [
        ResourceSummary(
            kind=DockerResourceKind.CONTAINER,
            id="1",
            name="api",
            raw={},
            columns={},
            group="atlas",
        ),
        ResourceSummary(
            kind=DockerResourceKind.CONTAINER,
            id="2",
            name="worker",
            raw={},
            columns={},
            group="odysseus",
        ),
        ResourceSummary(
            kind=DockerResourceKind.CONTAINER,
            id="3",
            name="zeta",
            raw={},
            columns={},
            group=None,
        ),
    ]

    table = build_resource_table(DockerResourceKind.CONTAINER, resources)

    assert table.display_rows[0] == table.resources[0]
    assert table.display_rows[1] is None
    assert table.display_rows[2] == table.resources[1]
    assert table.display_rows[3] is None
    assert table.display_rows[4] == table.resources[2]


def test_app_receives_docker_service():
    from main import Dokbox
    from models.config import DokboxConfig

    service = object()
    app = Dokbox(
        DokboxConfig(docker_host="unix:///tmp/docker.sock"),
        docker_service=service,
    )

    assert app.docker_service is service


def test_app_disables_textual_command_palette():
    from main import Dokbox

    assert Dokbox.ENABLE_COMMAND_PALETTE is False


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


def test_resource_browser_starts_with_tabs_focused():
    async def run_test():
        app = _BrowserApp()

        async with app.run_test() as pilot:
            await pilot.pause()
            shortcut_bar = app.screen_under_test.query_one(
                "#browser-shortcuts", ShortcutBar
            )

            assert isinstance(app.focused, ContentTabs)
            assert (
                shortcut_bar.renderable.plain
                == "←/→ Switch Tabs | ↵ Focus Table | q Quit"
            )

        return None

    asyncio.run(run_test())


def test_resource_browser_uses_single_command_bar():
    async def run_test():
        app = _BrowserApp()

        async with app.run_test() as pilot:
            await pilot.pause()

            assert len(app.screen_under_test.query(ShortcutBar)) == 1
            assert len(app.screen_under_test.query(Footer)) == 0

        return None

    asyncio.run(run_test())


def test_resource_browser_pins_command_bar_to_bottom():
    async def run_test():
        app = _BrowserApp()

        async with app.run_test() as pilot:
            await pilot.pause()
            shortcut_bar = app.screen_under_test.query_one(
                "#browser-shortcuts", ShortcutBar
            )

            assert str(shortcut_bar.styles.dock) == "bottom"

        return None

    asyncio.run(run_test())


def test_resource_browser_enters_table_on_enter_and_updates_shortcuts():
    async def run_test():
        app = _BrowserApp()

        async with app.run_test() as pilot:
            await pilot.pause()
            shortcut_bar = app.screen_under_test.query_one(
                "#browser-shortcuts", ShortcutBar
            )
            await pilot.press("enter")
            await pilot.pause()

            assert isinstance(app.focused, ResourceTable)
            assert app.focused.kind == DockerResourceKind.CONTAINER
            assert (
                shortcut_bar.renderable.plain
                == "↑/↓ Rows | ↵ Open Details | o Actions | p Prune | q Back To Tabs"
            )

        return None

    asyncio.run(run_test())


def test_resource_browser_returns_to_tabs_on_q():
    async def run_test():
        app = _BrowserApp()

        async with app.run_test() as pilot:
            await pilot.pause()
            await pilot.press("enter")
            await pilot.pause()
            await pilot.press("q")
            await pilot.pause()

            assert isinstance(app.focused, ContentTabs)

        return None

    asyncio.run(run_test())


def test_resource_browser_switches_tabs_before_entering_table():
    async def run_test():
        app = _BrowserApp()

        async with app.run_test() as pilot:
            await pilot.pause()
            await pilot.press("right")
            await pilot.pause()
            await pilot.press("enter")
            await pilot.pause()

            tabbed_content = app.screen_under_test.query_one(TabbedContent)

            assert tabbed_content.active == "images"
            assert isinstance(app.focused, ResourceTable)
            assert app.focused.kind == DockerResourceKind.IMAGE

        return None

    asyncio.run(run_test())


def test_resource_browser_opens_container_detail_screen_from_table():
    async def run_test():
        app = _BrowserApp()

        async with app.run_test() as pilot:
            await pilot.pause()
            await pilot.press("enter")
            await pilot.pause()
            await pilot.press("enter")
            await pilot.pause()

            from ui.screens import ContainerDetailScreen

            assert isinstance(app.screen, ContainerDetailScreen)
            assert app.screen.container_id == "container-1"

        return None

    asyncio.run(run_test())


def test_container_detail_screen_returns_to_browser_on_q():
    async def run_test():
        app = _BrowserApp()

        async with app.run_test() as pilot:
            await pilot.pause()
            await pilot.press("enter")
            await pilot.pause()
            await pilot.press("enter")
            await pilot.pause()
            await pilot.press("q")
            await pilot.pause()

            assert isinstance(app.screen, ResourceBrowserScreen)

        return None

    asyncio.run(run_test())


def test_resource_browser_opens_container_actions_modal_and_runs_action():
    async def run_test():
        app = _BrowserApp()

        async with app.run_test() as pilot:
            await pilot.pause()
            await pilot.press("enter")
            await pilot.pause()
            await pilot.press("o")
            await pilot.pause()

            from ui.screens import ContainerActionsDialog

            assert isinstance(app.screen, ContainerActionsDialog)

            await pilot.press("r")
            await pilot.pause()

            assert isinstance(app.screen, ResourceBrowserScreen)
            assert app.docker_service.performed_actions == [("restart", "container-1")]

        return None

    asyncio.run(run_test())


def test_container_actions_modal_uses_compact_overlay_container():
    async def run_test():
        app = _BrowserApp()

        async with app.run_test() as pilot:
            await pilot.pause()
            await pilot.press("enter")
            await pilot.pause()
            await pilot.press("o")
            await pilot.pause()

            from ui.screens import ContainerActionsDialog

            assert isinstance(app.screen, ContainerActionsDialog)
            direct_children = list(app.screen.children)

            assert len(direct_children) == 1
            assert isinstance(direct_children[0], Container)

        return None

    asyncio.run(run_test())


def test_container_actions_modal_renders_actions_directly_in_overlay_panel():
    async def run_test():
        app = _BrowserApp()

        async with app.run_test() as pilot:
            await pilot.pause()
            await pilot.press("enter")
            await pilot.pause()
            await pilot.press("o")
            await pilot.pause()

            from ui.screens import ContainerActionsDialog

            assert isinstance(app.screen, ContainerActionsDialog)
            panel = list(app.screen.children)[0]

            assert [child.__class__.__name__ for child in panel.children] == [
                "Label",
                "Static",
                "Static",
                "Static",
                "Static",
                "Static",
                "Static",
            ]

        return None

    asyncio.run(run_test())


def test_resource_browser_prunes_containers_from_table_shortcut():
    async def run_test():
        app = _BrowserApp()

        async with app.run_test() as pilot:
            await pilot.pause()
            await pilot.press("enter")
            await pilot.pause()
            await pilot.press("p")
            await pilot.pause()

            assert app.docker_service.performed_actions == [
                ("prune", DockerResourceKind.CONTAINER)
            ]

        return None

    asyncio.run(run_test())


def test_container_actions_modal_supports_prune():
    async def run_test():
        app = _BrowserApp()

        async with app.run_test() as pilot:
            await pilot.pause()
            await pilot.press("enter")
            await pilot.pause()
            await pilot.press("o")
            await pilot.pause()
            await pilot.press("p")
            await pilot.pause()

            from ui.screens import ResourceBrowserScreen

            assert isinstance(app.screen, ResourceBrowserScreen)
            assert app.docker_service.performed_actions == [
                ("prune", DockerResourceKind.CONTAINER)
            ]

        return None

    asyncio.run(run_test())
