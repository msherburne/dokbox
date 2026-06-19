import sys
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


def test_resource_summary_exposes_stable_table_identity():
    summary = ResourceSummary(
        kind=DockerResourceKind.CONTAINER,
        id="abcdef1234567890",
        name="api",
        raw={},
        columns={"Image": "dokbox-api", "State": "running"},
    )

    assert summary.short_id == "abcdef123456"
    assert summary.columns["Image"] == "dokbox-api"
    assert summary.kind == DockerResourceKind.CONTAINER


def test_container_columns_and_row_are_stable():
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

    assert get_resource_columns(DockerResourceKind.CONTAINER) == [
        "Name",
        "Image",
        "State",
        "Status",
        "Ports",
        "Created",
    ]
    assert resource_to_row(summary) == [
        "api",
        "dokbox-api",
        "running",
        "Up 3 minutes",
        "8080->80",
        "2026-06-19 04:00",
    ]


def test_shortcuts_are_contextual():
    shortcuts = get_shortcuts("containers")

    assert ShortcutHint("Enter", "Details") in shortcuts
    assert ShortcutHint("p", "Prune") in shortcuts
    assert ShortcutHint("Backspace", "Up") not in shortcuts


def test_metric_card_uses_bar_for_limited_metric():
    metric = MetricSample(
        name="CPU", value=0.42, limit=1.0, unit="cores", label="42% of 1.0 core"
    )

    card = build_metric_card(metric)

    assert card.title == "CPU"
    assert card.value == "42% of 1.0 core"
    assert card.bar == "####------ 42%"


def test_prune_shortcut_exists_for_each_resource_tab():
    for context in ["containers", "images", "volumes", "networks"]:
        assert ShortcutHint("p", "Prune") in get_shortcuts(context)
