import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "src"))

from models.docker_resources import DockerResourceKind, ResourceSummary


class ResourceViewModelTest(unittest.TestCase):
    def test_resource_summary_exposes_stable_table_identity(self):
        summary = ResourceSummary(
            kind=DockerResourceKind.CONTAINER,
            id="abcdef1234567890",
            name="api",
            raw={},
            columns={"Image": "dokbox-api", "State": "running"},
        )

        self.assertEqual(summary.short_id, "abcdef123456")
        self.assertEqual(summary.columns["Image"], "dokbox-api")
        self.assertEqual(summary.kind, DockerResourceKind.CONTAINER)


if __name__ == "__main__":
    unittest.main()
