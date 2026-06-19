from dokbox.models.config import DokboxConfig
from dokbox.utils.docker import detect_docker_host

DEFAULT_CONFIG = DokboxConfig(docker_host="")


def generate_default_config() -> DokboxConfig:
    socket_address = detect_docker_host()
    DEFAULT_CONFIG.docker_host = socket_address

    print(f"Using default configuration:\n{DEFAULT_CONFIG.model_dump_json(indent=2)}")
    return DEFAULT_CONFIG
