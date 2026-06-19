import os
import sys
import tempfile
from pathlib import Path
from unittest.mock import patch

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "src"))

from models.config import DokboxConfig
from utils.setup import ask_for_config_save, load_or_setup_config


def test_ask_for_config_save_writes_model_dump():
    with tempfile.TemporaryDirectory() as home:
        config = DokboxConfig(docker_host="unix:///tmp/docker.sock")

        with (
            patch.dict(os.environ, {"HOME": home}),
            patch("utils.setup.get_confirmation", return_value=True),
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
