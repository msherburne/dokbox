from textual.app import ComposeResult
from textual.containers import Center, Middle, Vertical
from textual.screen import ModalScreen, Screen
from textual.widgets import (
    Button,
    Footer,
    Header,
    Label,
    Static,
    TabbedContent,
    TabPane,
)

from models.docker_resources import DockerResourceKind
from services.docker_service import DockerService
from ui.widgets import ShortcutBar, build_resource_table
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


class ResourceBrowserScreen(Screen):
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
        with TabbedContent(initial="containers"):
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
        yield ShortcutBar(get_shortcuts("containers"))
        yield Footer()


class ContainerDetailScreen(Screen):
    def __init__(self, docker_service: DockerService, container_id: str):
        super().__init__()
        self.docker_service = docker_service
        self.container_id = container_id

    def compose(self) -> ComposeResult:
        yield Header(show_clock=True)
        with TabbedContent(initial="overview"):
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
        yield ShortcutBar(get_shortcuts("containers"))
        yield Footer()

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
