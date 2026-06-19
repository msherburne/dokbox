from textual.app import App
from utils.setup import load_or_setup_config
from models.config import DokboxConfig
from utils.docker import load_docker_client


class Dokbox(App):
    def __init__(self, config: DokboxConfig):
        super().__init__()
        self.config = config
        self.docker_client = load_docker_client(config.docker_host)


if __name__ == "__main__":
    config = load_or_setup_config()
    app = Dokbox(config)
    app.run()
