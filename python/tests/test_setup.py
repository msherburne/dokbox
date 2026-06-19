import os
import tempfile
from importlib import import_module
from pathlib import Path
from unittest.mock import patch

from dokbox.models.config import DokboxConfig
from dokbox.utils.setup import ask_for_config_save, load_or_setup_config

main = import_module("dokbox.main")
setup = import_module("dokbox.utils.setup")


def test_ask_for_config_save_writes_model_dump():
    with tempfile.TemporaryDirectory() as home:
        config = DokboxConfig(docker_host="unix:///tmp/docker.sock")

        with (
            patch.dict(os.environ, {"HOME": home}),
            patch.object(setup, "get_confirmation", return_value=True),
        ):
            ask_for_config_save(config)

        saved = (Path(home) / ".dokbox.json").read_text()

    assert '"docker_host": "unix:///tmp/docker.sock"' in saved
    assert '"config_version": 1' in saved


def test_load_or_setup_config_resolves_existing_config():
    with tempfile.TemporaryDirectory() as home:
        config_path = Path(home) / ".dokbox.json"
        config_path.write_text('{"docker_host": "unix:///tmp/docker.sock"}')

        with patch.dict(os.environ, {"HOME": home}):
            config = load_or_setup_config()

    assert config.config_version == 1
    assert config.docker_host == "unix:///tmp/docker.sock"


def test_main_uses_injected_docker_service_without_loading_client(monkeypatch):
    config = DokboxConfig(docker_host="unix:///tmp/docker.sock")
    docker_service: list[object] = []
    loaded_hosts: list[str] = []

    monkeypatch.setattr(
        main, "load_docker_client", lambda docker_host: loaded_hosts.append(docker_host)
    )

    app = main.Dokbox(config, docker_service=docker_service)

    assert app.docker_client is None
    assert app.docker_service is docker_service
    assert loaded_hosts == []
