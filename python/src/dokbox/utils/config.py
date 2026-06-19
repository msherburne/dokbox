import json
import os

from dokbox.models.config import DokboxConfig

CURRENT_CONFIG_VERSION = 1


def get_config_path() -> str:
    return os.path.expanduser("~/.dokbox.json")


def save_config(config: DokboxConfig) -> None:
    with open(get_config_path(), "w") as f:
        json.dump(config.model_dump(), f, indent=2)


def load_raw_config() -> dict | None:
    config_path = get_config_path()
    if not os.path.exists(config_path):
        return None

    with open(config_path, "r") as f:
        return json.load(f)


def load_config() -> DokboxConfig | None:
    config = load_raw_config()
    if config is None:
        return None
    return DokboxConfig(**config)


def migrate_config(config: dict) -> dict:
    migrated = dict(config)
    migrated.setdefault("config_version", CURRENT_CONFIG_VERSION)
    return migrated


def resolve_current_config() -> DokboxConfig | None:
    raw_config = load_raw_config()
    if raw_config is None:
        return None

    config = DokboxConfig(**migrate_config(raw_config))
    save_config(config)
    return config
