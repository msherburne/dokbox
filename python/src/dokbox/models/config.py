from pydantic import BaseModel


class DokboxConfig(BaseModel):
    config_version: int = 1
    docker_host: str
