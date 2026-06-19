import os
import tempfile
from pathlib import Path
from unittest.mock import patch

from models.config import DokboxConfig
from utils.config import (
    load_config,
    load_raw_config,
    resolve_current_config,
    save_config,
)


def test_config_has_current_version_by_default():
    config = DokboxConfig(docker_host="unix:///var/run/docker.sock")

    assert config.config_version == 1


def test_load_raw_config_returns_dict_without_validation():
    with tempfile.TemporaryDirectory() as home:
        config_path = Path(home) / ".dokbox.json"
        config_path.write_text('{"docker_host": "unix:///tmp/docker.sock"}')

        with patch.dict(os.environ, {"HOME": home}):
            raw_config = load_raw_config()

    assert raw_config == {"docker_host": "unix:///tmp/docker.sock"}


def test_load_config_validates_current_config():
    with tempfile.TemporaryDirectory() as home:
        config = DokboxConfig(docker_host="unix:///tmp/docker.sock")

        with patch.dict(os.environ, {"HOME": home}):
            save_config(config)
            loaded_config = load_config()

    assert loaded_config == config


def test_resolve_current_config_adds_missing_version_and_saves():
    with tempfile.TemporaryDirectory() as home:
        config_path = Path(home) / ".dokbox.json"
        config_path.write_text('{"docker_host": "unix:///tmp/docker.sock"}')

        with patch.dict(os.environ, {"HOME": home}):
            config = resolve_current_config()
            raw_config = load_raw_config()

    assert config.config_version == 1
    assert config.docker_host == "unix:///tmp/docker.sock"
    assert raw_config["config_version"] == 1


def test_resolve_current_config_returns_none_when_missing():
    with tempfile.TemporaryDirectory() as home:
        with patch.dict(os.environ, {"HOME": home}):
            config = resolve_current_config()

    assert config is None
