import os
import sys
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "src"))

from models.config import DokboxConfig
from utils.setup import ask_for_config_save, load_or_setup_config


class SetupConfigTest(unittest.TestCase):
    def test_ask_for_config_save_writes_model_dump(self):
        with tempfile.TemporaryDirectory() as home:
            config = DokboxConfig(docker_host="unix:///tmp/docker.sock")

            with (
                patch.dict(os.environ, {"HOME": home}),
                patch("utils.setup.get_confirmation", return_value=True),
            ):
                ask_for_config_save(config)

            saved = (Path(home) / ".dokbox.json").read_text()

        self.assertIn('"docker_host": "unix:///tmp/docker.sock"', saved)
        self.assertIn('"config_version": 1', saved)

    def test_load_or_setup_config_resolves_existing_config(self):
        with tempfile.TemporaryDirectory() as home:
            config_path = Path(home) / ".dokbox.json"
            config_path.write_text('{"docker_host": "unix:///tmp/docker.sock"}')

            with patch.dict(os.environ, {"HOME": home}):
                config = load_or_setup_config()

        self.assertEqual(config.config_version, 1)
        self.assertEqual(config.docker_host, "unix:///tmp/docker.sock")


if __name__ == "__main__":
    unittest.main()
