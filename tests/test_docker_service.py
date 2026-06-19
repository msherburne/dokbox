import sys
import unittest
from pathlib import Path
from unittest.mock import Mock, call

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

    def test_container_lifecycle_actions_route_to_selected_container(self):
        container = Mock()
        client = Mock()
        client.containers.get.return_value = container
        service = DockerService(client)

        service.start_container("abc")
        service.stop_container("abc")
        service.restart_container("abc")
        service.remove_resource(DockerResourceKind.CONTAINER, "abc")

        self.assertEqual(
            client.containers.get.call_args_list,
            [call("abc"), call("abc"), call("abc"), call("abc")],
        )
        container.start.assert_called_once()
        container.stop.assert_called_once()
        container.restart.assert_called_once()
        container.remove.assert_called_once()

    def test_remove_image_routes_to_images_remove(self):
        client = Mock()
        service = DockerService(client)

        service.remove_resource(DockerResourceKind.IMAGE, "img")

        client.images.remove.assert_called_once_with("img")

    def test_remove_volume_routes_to_volume_remove(self):
        volume = Mock()
        client = Mock()
        client.volumes.get.return_value = volume
        service = DockerService(client)

        service.remove_resource(DockerResourceKind.VOLUME, "vol")

        client.volumes.get.assert_called_once_with("vol")
        volume.remove.assert_called_once()

    def test_remove_network_routes_to_network_remove(self):
        network = Mock()
        client = Mock()
        client.networks.get.return_value = network
        service = DockerService(client)

        service.remove_resource(DockerResourceKind.NETWORK, "net")

        client.networks.get.assert_called_once_with("net")
        network.remove.assert_called_once()

    def test_connection_check_returns_connected_after_ping(self):
        client = Mock()
        service = DockerService(client)

        result = service.check_connection()

        client.ping.assert_called_once()
        self.assertTrue(result.ok)
        self.assertEqual(result.message, "Connected")

    def test_connection_check_returns_message_instead_of_raising(self):
        client = Mock()
        client.ping.side_effect = RuntimeError("cannot connect")
        service = DockerService(client)

        result = service.check_connection()

        self.assertFalse(result.ok)
        self.assertEqual(result.message, "cannot connect")


if __name__ == "__main__":
    unittest.main()
