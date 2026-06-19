import os
import sys
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "src"))

from models.config import DokboxConfig
from utils.config import (
    load_config,
    load_raw_config,
    resolve_current_config,
    save_config,
)


class DokboxConfigTest(unittest.TestCase):
    def test_config_has_current_version_by_default(self):
        config = DokboxConfig(docker_host="unix:///var/run/docker.sock")

        self.assertEqual(config.config_version, 1)

    def test_load_raw_config_returns_dict_without_validation(self):
        with tempfile.TemporaryDirectory() as home:
            config_path = Path(home) / ".dokbox.json"
            config_path.write_text('{"docker_host": "unix:///tmp/docker.sock"}')

            with patch.dict(os.environ, {"HOME": home}):
                raw_config = load_raw_config()

        self.assertEqual(raw_config, {"docker_host": "unix:///tmp/docker.sock"})

    def test_load_config_validates_current_config(self):
        with tempfile.TemporaryDirectory() as home:
            config = DokboxConfig(docker_host="unix:///tmp/docker.sock")

            with patch.dict(os.environ, {"HOME": home}):
                save_config(config)
                loaded_config = load_config()

        self.assertEqual(loaded_config, config)

    def test_resolve_current_config_adds_missing_version_and_saves(self):
        with tempfile.TemporaryDirectory() as home:
            config_path = Path(home) / ".dokbox.json"
            config_path.write_text('{"docker_host": "unix:///tmp/docker.sock"}')

            with patch.dict(os.environ, {"HOME": home}):
                config = resolve_current_config()
                raw_config = load_raw_config()

        self.assertEqual(config.config_version, 1)
        self.assertEqual(config.docker_host, "unix:///tmp/docker.sock")
        self.assertEqual(raw_config["config_version"], 1)

    def test_resolve_current_config_returns_none_when_missing(self):
        with tempfile.TemporaryDirectory() as home:
            with patch.dict(os.environ, {"HOME": home}):
                config = resolve_current_config()

        self.assertIsNone(config)


if __name__ == "__main__":
    unittest.main()
