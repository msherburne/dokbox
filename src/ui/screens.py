from textual.app import ComposeResult
from textual.containers import Center, Middle, Vertical
from textual.screen import ModalScreen, Screen
from textual.widgets import Button, Footer, Header, Label, TabbedContent, TabPane

from models.docker_resources import DockerResourceKind
from services.docker_service import DockerService
from ui.widgets import ShortcutBar, build_resource_table
from view_models.resources import get_shortcuts


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
