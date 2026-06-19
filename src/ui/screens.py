from textual.app import ComposeResult
from textual.containers import Center, Container, Middle, Vertical
from textual.dom import NoMatches
from textual.screen import ModalScreen, Screen
from textual.widgets import (
    Button,
    Header,
    Label,
    Static,
    TabbedContent,
    TabPane,
)
from textual.widgets._tabbed_content import ContentTabs

from models.docker_resources import DockerResourceKind
from services.docker_service import DockerService
from ui.widgets import ResourceTable, ShortcutBar, build_resource_table
from view_models.resources import build_metric_card, get_shortcuts


class ConfirmActionDialog(ModalScreen[bool]):
    def __init__(self, message: str):
        super().__init__()
        self.message = message

    def compose(self) -> ComposeResult:
        with Center():
            with Middle():
                yield Label(self.message)
                yield Button("Cancel", id="cancel")
                yield Button("Confirm", id="confirm", variant="error")

    def on_button_pressed(self, event: Button.Pressed) -> None:
        self.dismiss(event.button.id == "confirm")


class ContainerActionsDialog(ModalScreen[str | None]):
    DEFAULT_CSS = """
    ContainerActionsDialog {
        align: center middle;
    }

    ContainerActionsDialog > .actions-dialog {
        width: 28;
        height: auto;
        padding: 1 2;
        border: round $primary;
        background: $surface;
    }
    """

    BINDINGS = [
        ("s", "run_action('start')", "Start"),
        ("t", "run_action('stop')", "Stop"),
        ("r", "run_action('restart')", "Restart"),
        ("x", "run_action('remove')", "Remove"),
        ("p", "run_action('prune')", "Prune"),
        ("q", "cancel", "Cancel"),
        ("escape", "cancel", "Cancel"),
    ]

    def __init__(self, container_name: str):
        super().__init__()
        self.container_name = container_name

    def compose(self) -> ComposeResult:
        with Container(classes="actions-dialog"):
            yield Label(f"Actions: {self.container_name}")
            yield Static("s Start")
            yield Static("t Stop")
            yield Static("r Restart")
            yield Static("x Remove")
            yield Static("p Prune")
            yield Static("q Cancel")

    def action_run_action(self, action_name: str) -> None:
        self.dismiss(action_name)

    def action_cancel(self) -> None:
        self.dismiss(None)


class ResourceBrowserScreen(Screen):
    DEFAULT_CSS = """
    ResourceBrowserScreen {
        layout: vertical;
        overflow: hidden;
    }

    #browser-tabs {
        height: 1fr;
    }

    #browser-shortcuts {
        dock: bottom;
    }
    """

    BINDINGS = [
        ("enter", "enter_table", "Focus Table"),
        ("o", "open_actions", "Actions"),
        ("p", "prune_active", "Prune"),
        ("q", "back_or_quit", "Back"),
    ]

    def __init__(self, docker_service: DockerService):
        super().__init__()
        self.docker_service = docker_service

    def compose(self) -> ComposeResult:
        yield Header(show_clock=True)
        status = self.docker_service.check_connection()
        if not status.ok:
            yield Label(f"Docker connection failed: {status.message}")
            yield ShortcutBar(get_shortcuts("containers"))
            return
        with TabbedContent(initial="containers", id="browser-tabs"):
            for kind, tab_id, title in [
                (DockerResourceKind.CONTAINER, "containers", "Containers"),
                (DockerResourceKind.IMAGE, "images", "Images"),
                (DockerResourceKind.VOLUME, "volumes", "Volumes"),
                (DockerResourceKind.NETWORK, "networks", "Networks"),
            ]:
                with TabPane(title, id=tab_id):
                    with Vertical():
                        yield build_resource_table(
                            kind, self.docker_service.list_resources(kind)
                        )
        yield ShortcutBar(get_shortcuts("resource-tabs"), id="browser-shortcuts")

    def on_mount(self) -> None:
        self.call_after_refresh(self._focus_tabs)

    def on_tabbed_content_tab_activated(
        self, event: TabbedContent.TabActivated
    ) -> None:
        self._refresh_shortcuts("resource-tabs")

    def action_enter_table(self) -> None:
        if isinstance(self.app.focused, ContentTabs):
            self._focus_active_table()

    def on_data_table_row_selected(self, event: ResourceTable.RowSelected) -> None:
        table = event.data_table
        if not isinstance(table, ResourceTable):
            return
        self._open_selected_resource(table)

    def action_back_or_quit(self) -> None:
        if isinstance(self.app.focused, ResourceTable):
            self._focus_tabs()
            return
        self.app.exit()

    def action_open_actions(self) -> None:
        table = self.app.focused
        if (
            not isinstance(table, ResourceTable)
            or table.kind != DockerResourceKind.CONTAINER
        ):
            return
        selected = table.selected_resource
        if selected is None:
            return
        self.app.push_screen(
            ContainerActionsDialog(selected.name),
            callback=lambda action: self._run_container_action(
                selected.id, table.kind, action
            ),
        )

    def action_prune_active(self) -> None:
        table = self.app.focused
        if not isinstance(table, ResourceTable):
            return
        self.docker_service.prune(table.kind)
        self._refresh_active_table()

    def _focus_active_table(self) -> None:
        try:
            active_pane = self.query_one(TabbedContent).active
            table = self.query_one(f"#{active_pane} ResourceTable", ResourceTable)
        except NoMatches:
            return
        table.focus()
        self._refresh_shortcuts(f"{active_pane}-table")

    def _focus_tabs(self) -> None:
        try:
            tabs = self.query_one(ContentTabs)
        except NoMatches:
            return
        tabs.focus()
        self._refresh_shortcuts("resource-tabs")

    def _refresh_shortcuts(self, context: str) -> None:
        try:
            shortcut_bar = self.query_one("#browser-shortcuts", ShortcutBar)
        except NoMatches:
            return
        shortcut_bar.set_shortcuts(get_shortcuts(context))

    def _open_selected_resource(self, table: ResourceTable) -> None:
        if not isinstance(table, ResourceTable):
            return
        selected = table.selected_resource
        if selected is None:
            return
        if selected.kind == DockerResourceKind.CONTAINER:
            self.app.push_screen(
                ContainerDetailScreen(self.docker_service, selected.id)
            )

    def _run_container_action(
        self,
        container_id: str,
        kind: DockerResourceKind,
        action_name: str | None,
    ) -> None:
        if action_name == "start":
            self.docker_service.start_container(container_id)
        elif action_name == "stop":
            self.docker_service.stop_container(container_id)
        elif action_name == "restart":
            self.docker_service.restart_container(container_id)
        elif action_name == "remove":
            self.docker_service.remove_resource(
                DockerResourceKind.CONTAINER, container_id
            )
        elif action_name == "prune":
            self.docker_service.prune(kind)
        else:
            return
        self._refresh_active_table()

    def _refresh_active_table(self) -> None:
        try:
            active_pane = self.query_one(TabbedContent).active
            table = self.query_one(f"#{active_pane} ResourceTable", ResourceTable)
        except NoMatches:
            return
        kind = table.kind
        table.refresh_resources(self.docker_service.list_resources(kind))


class ContainerDetailScreen(Screen):
    DEFAULT_CSS = """
    ContainerDetailScreen {
        layout: vertical;
        overflow: hidden;
    }

    #detail-tabs {
        height: 1fr;
    }

    #detail-shortcuts {
        dock: bottom;
    }
    """

    BINDINGS = [("q", "back", "Back")]

    def __init__(self, docker_service: DockerService, container_id: str):
        super().__init__()
        self.docker_service = docker_service
        self.container_id = container_id

    def compose(self) -> ComposeResult:
        yield Header(show_clock=True)
        with TabbedContent(initial="overview", id="detail-tabs"):
            with TabPane("Overview", id="overview"):
                for metric in self.docker_service.get_container_metrics(
                    self.container_id
                ):
                    card = build_metric_card(metric)
                    yield Static(f"{card.title}: {card.value}\n{card.bar}")
            with TabPane("Logs", id="logs"):
                yield LogsPane(self.docker_service, self.container_id)
            with TabPane("Shell", id="shell"):
                yield ShellPane(self.docker_service, self.container_id)
            with TabPane("Files", id="files"):
                yield FilesPane(self.docker_service, self.container_id, "/")
        yield ShortcutBar(get_shortcuts("container-details"), id="detail-shortcuts")

    def action_back(self) -> None:
        self.app.pop_screen()


class LogsPane(Static):
    def __init__(self, docker_service: DockerService, container_id: str):
        self.docker_service = docker_service
        self.container_id = container_id
        super().__init__("Loading logs...")

    def on_mount(self) -> None:
        lines = []
        for index, line in enumerate(
            self.docker_service.stream_logs(self.container_id, follow=False)
        ):
            if index >= 200:
                break
            lines.append(line.text)
        self.update("\n".join(lines) if lines else "No logs.")


class ShellPane(Static):
    def __init__(self, docker_service: DockerService, container_id: str):
        self.docker_service = docker_service
        self.container_id = container_id
        super().__init__("Press Enter to start shell.")


class FilesPane(Static):
    def __init__(
        self, docker_service: DockerService, container_id: str, path: str = "/"
    ):
        self.docker_service = docker_service
        self.container_id = container_id
        self.path = path
        super().__init__("Loading files...")

    def on_mount(self) -> None:
        entries = self.docker_service.list_container_path(self.container_id, self.path)
        rendered = "\n".join(
            f"{'[d]' if entry.is_dir else '[f]'} {entry.name} {entry.mode}"
            for entry in entries
        )
        self.update(rendered if rendered else "No readable entries.")
