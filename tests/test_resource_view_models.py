import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "src"))

from models.docker_resources import DockerResourceKind, MetricSample, ResourceSummary
from view_models.resources import (
    ShortcutHint,
    build_metric_card,
    get_resource_columns,
    get_shortcuts,
    resource_to_row,
)


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

    def test_container_columns_and_row_are_stable(self):
        summary = ResourceSummary(
            kind=DockerResourceKind.CONTAINER,
            id="abcdef1234567890",
            name="api",
            raw={},
            columns={
                "Image": "dokbox-api",
                "State": "running",
                "Status": "Up 3 minutes",
                "Ports": "8080->80",
                "Created": "2026-06-19 04:00",
            },
        )

        self.assertEqual(
            get_resource_columns(DockerResourceKind.CONTAINER),
            ["Name", "Image", "State", "Status", "Ports", "Created"],
        )
        self.assertEqual(
            resource_to_row(summary),
            [
                "api",
                "dokbox-api",
                "running",
                "Up 3 minutes",
                "8080->80",
                "2026-06-19 04:00",
            ],
        )

    def test_shortcuts_are_contextual(self):
        shortcuts = get_shortcuts("containers")

        self.assertIn(ShortcutHint("Enter", "Details"), shortcuts)
        self.assertIn(ShortcutHint("p", "Prune"), shortcuts)
        self.assertNotIn(ShortcutHint("Backspace", "Up"), shortcuts)

    def test_metric_card_uses_bar_for_limited_metric(self):
        metric = MetricSample(
            name="CPU", value=0.42, limit=1.0, unit="cores", label="42% of 1.0 core"
        )

        card = build_metric_card(metric)

        self.assertEqual(card.title, "CPU")
        self.assertEqual(card.value, "42% of 1.0 core")
        self.assertEqual(card.bar, "####------ 42%")

    def test_prune_shortcut_exists_for_each_resource_tab(self):
        for context in ["containers", "images", "volumes", "networks"]:
            self.assertIn(ShortcutHint("p", "Prune"), get_shortcuts(context))


if __name__ == "__main__":
    unittest.main()
