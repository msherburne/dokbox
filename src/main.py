from textual.app import App
from models.config import DokboxConfig
from services.docker_service import DockerService
from ui.screens import ResourceBrowserScreen
from utils.docker import load_docker_client
from utils.setup import load_or_setup_config


class Dokbox(App):
    def __init__(
        self, config: DokboxConfig, docker_service: DockerService | None = None
    ):
        super().__init__()
        self.config = config
        self.docker_client = (
            load_docker_client(config.docker_host) if docker_service is None else None
        )
        self.docker_service = docker_service or DockerService(self.docker_client)

    def on_mount(self) -> None:
        self.push_screen(ResourceBrowserScreen(self.docker_service))


if __name__ == "__main__":
    config = load_or_setup_config()
    app = Dokbox(config)
    app.run()
