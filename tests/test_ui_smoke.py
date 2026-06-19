import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "src"))

from models.docker_resources import DockerResourceKind, ResourceSummary
from ui.widgets import ResourceTable, ShortcutBar, build_resource_table
from view_models.resources import ShortcutHint


class UiSmokeTest(unittest.TestCase):
    def test_shortcut_bar_renders_contextual_hints(self):
        bar = ShortcutBar([ShortcutHint("Enter", "Details"), ShortcutHint("q", "Quit")])

        self.assertEqual(bar.renderable, "Enter Details  q Quit")

    def test_build_resource_table_stores_columns_and_rows_for_mount(self):
        resources = [
            ResourceSummary(
                kind=DockerResourceKind.CONTAINER,
                id="abcdef1234567890",
                name="api",
                raw={},
                columns={
                    "Image": "dokbox-api",
                    "State": "running",
                    "Status": "Up",
                    "Ports": "8080->80",
                    "Created": "-",
                },
            )
        ]

        table = build_resource_table(DockerResourceKind.CONTAINER, resources)

        self.assertIsInstance(table, ResourceTable)
        self.assertEqual(table.cursor_type, "row")
        self.assertEqual(table.kind, DockerResourceKind.CONTAINER)
        self.assertEqual(table.resources, resources)


if __name__ == "__main__":
    unittest.main()
