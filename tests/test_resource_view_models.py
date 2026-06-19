import sys
from pathlib import Path

from rich.text import Text

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "src"))

from models.docker_resources import DockerResourceKind, MetricSample, ResourceSummary
from view_models.resources import (
    PaletteEntry,
    ShortcutHint,
    build_container_palette_entries,
    build_metric_card,
    container_cell,
    filter_palette_entries,
    get_resource_columns,
    get_shortcuts,
    resource_to_row,
    stack_label,
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
        "Stack",
        "Name",
        "Image",
        "State",
        "Status",
        "Ports",
        "Created",
    ]
    row = resource_to_row(summary)
    assert isinstance(row[0], Text)
    assert row[0].plain == "Ungrouped"
    assert [cell.plain for cell in row[1:]] == [
        "api",
        "dokbox-api",
        "running",
        "Up 3 minutes",
        "8080->80",
        "2026-06-19 04:00",
    ]


def test_stack_label_uses_consistent_color_for_same_group():
    summary = ResourceSummary(
        kind=DockerResourceKind.CONTAINER,
        id="a1",
        name="api",
        raw={},
        columns={"Image": "api:latest"},
        group="odysseus",
    )

    label = stack_label(summary)

    assert isinstance(label, Text)
    assert label.plain == "odysseus"
    assert str(label.style)


def test_container_cells_dim_when_container_is_not_running():
    summary = ResourceSummary(
        kind=DockerResourceKind.CONTAINER,
        id="a1",
        name="api",
        raw={},
        columns={"State": "exited"},
        group="odysseus",
    )

    cell = container_cell(summary, "api")

    assert isinstance(cell, Text)
    assert cell.plain == "api"
    assert str(cell.style) == "dim"


def test_shortcuts_are_contextual():
    shortcuts = get_shortcuts("containers-table")

    assert ShortcutHint("Enter", "Open Details") in shortcuts
    assert ShortcutHint("o", "Actions") in shortcuts
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
    for context in [
        "containers-table",
        "images-table",
        "volumes-table",
        "networks-table",
    ]:
        assert ShortcutHint("p", "Prune") in get_shortcuts(context)


def test_resource_tab_shortcuts_show_navigation():
    shortcuts = get_shortcuts("resource-tabs")

    assert ShortcutHint("Left/Right", "Switch Tabs") in shortcuts
    assert ShortcutHint("Enter", "Focus Table") in shortcuts
    assert ShortcutHint("q", "Quit") in shortcuts


def test_container_detail_shortcuts_show_back_navigation():
    shortcuts = get_shortcuts("container-details")

    assert ShortcutHint("Left/Right", "Switch Pane") in shortcuts
    assert ShortcutHint("q", "Back") in shortcuts


def test_container_palette_entries_include_searchable_container_context():
    resources = [
        ResourceSummary(
            kind=DockerResourceKind.CONTAINER,
            id="abc123456789",
            name="api",
            raw={},
            columns={
                "Image": "dokbox-api:latest",
                "State": "running",
                "Status": "Up 2 hours",
            },
            group="atlas",
        )
    ]

    entries = build_container_palette_entries(resources)

    assert entries == [
        PaletteEntry(
            kind=DockerResourceKind.CONTAINER,
            resource_id="abc123456789",
            title="api",
            subtitle="atlas | dokbox-api:latest | Up 2 hours",
            search_text="api atlas dokbox-api:latest running up 2 hours",
        )
    ]


def test_container_palette_filter_matches_name_stack_image_and_status():
    entries = [
        PaletteEntry(
            kind=DockerResourceKind.CONTAINER,
            resource_id="1",
            title="api",
            subtitle="atlas | dokbox-api:latest | Up 2 hours",
            search_text="api atlas dokbox-api:latest running up 2 hours",
        ),
        PaletteEntry(
            kind=DockerResourceKind.CONTAINER,
            resource_id="2",
            title="worker",
            subtitle="odysseus | dokbox-worker:latest | Exited",
            search_text="worker odysseus dokbox-worker:latest exited",
        ),
    ]

    assert [entry.title for entry in filter_palette_entries(entries, "atlas")] == [
        "api"
    ]
    assert [entry.title for entry in filter_palette_entries(entries, "worker")] == [
        "worker"
    ]
    assert [entry.title for entry in filter_palette_entries(entries, "exited")] == [
        "worker"
    ]
    assert [entry.title for entry in filter_palette_entries(entries, "dokbox-api")] == [
        "api"
    ]
    assert [
        entry.title for entry in filter_palette_entries(entries, "worker exited")
    ] == ["worker"]
