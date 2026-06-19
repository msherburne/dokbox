import sys
import unittest
from pathlib import Path
from unittest.mock import Mock

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "src"))

from models.docker_resources import DockerResourceKind
from services.docker_service import DockerService


class ContainerStub:
    id = "abcdef1234567890"
    name = "api"
    status = "running"
    attrs = {
        "Config": {"Image": "dokbox-api:latest"},
        "Image": "sha256:1234",
        "State": {"Status": "running", "Health": {"Status": "healthy"}},
        "NetworkSettings": {"Ports": {"80/tcp": [{"HostPort": "8080"}]}},
        "Created": "2026-06-19T04:00:00Z",
    }

    @property
    def image(self):
        raise AssertionError("container.image should not be accessed")


class DockerServiceTest(unittest.TestCase):
    def test_list_containers_returns_resource_summaries(self):
        client = Mock()
        client.containers.list.return_value = [ContainerStub()]
        service = DockerService(client)

        summaries = service.list_resources(DockerResourceKind.CONTAINER)

        self.assertEqual(len(summaries), 1)
        self.assertEqual(summaries[0].name, "api")
        self.assertEqual(summaries[0].columns["Image"], "dokbox-api:latest")
        self.assertEqual(summaries[0].columns["Ports"], "8080->80")

    def test_prune_routes_by_resource_kind(self):
        client = Mock()
        service = DockerService(client)

        service.prune(DockerResourceKind.IMAGE)

        client.images.prune.assert_called_once()
        client.containers.prune.assert_not_called()


if __name__ == "__main__":
    unittest.main()
