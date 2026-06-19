from textual.app import App

from dokbox.models.config import DokboxConfig
from dokbox.services.docker_service import DockerService
from dokbox.ui.screens import ResourceBrowserScreen
from dokbox.utils.docker import load_docker_client
from dokbox.utils.setup import load_or_setup_config


class Dokbox(App):
    ENABLE_COMMAND_PALETTE = False

    def __init__(
        self, config: DokboxConfig, docker_service: DockerService | None = None
    ):
        super().__init__()
        self.config = config
        if docker_service is None:
            self.docker_client = load_docker_client(config.docker_host)
            self.docker_service = DockerService(self.docker_client)
        else:
            self.docker_client = None
            self.docker_service = docker_service

    def on_mount(self) -> None:
        self.push_screen(ResourceBrowserScreen(self.docker_service))


def main() -> None:
    config = load_or_setup_config()
    app = Dokbox(config)
    app.run()
