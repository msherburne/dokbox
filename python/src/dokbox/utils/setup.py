from dokbox.defaults.config import generate_default_config
from dokbox.models.config import DokboxConfig
from dokbox.utils.cli import get_confirmation
from dokbox.utils.config import resolve_current_config, save_config


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
