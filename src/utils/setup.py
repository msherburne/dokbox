from utils.cli import get_confirmation
from defaults.config import generate_default_config
from utils.config import resolve_current_config, save_config
from models.config import DokboxConfig


def ask_for_config_save(config: DokboxConfig):
    confirmation = get_confirmation("Do you want to save this configuration?")
    if confirmation:
        save_config(config)


def load_or_setup_config() -> DokboxConfig:
    config = resolve_current_config()
    if not config:
        config = generate_default_config()
        ask_for_config_save(config)
    return config
