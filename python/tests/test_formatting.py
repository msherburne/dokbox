from dokbox.view_models.formatting import (
    format_bytes,
    format_ratio_bar,
    format_timestamp,
)


def test_format_bytes_uses_binary_units():
    assert format_bytes(512) == "512 B"
    assert format_bytes(1536) == "1.5 KiB"
    assert format_bytes(1048576) == "1.0 MiB"


def test_format_ratio_bar_clamps_values():
    assert format_ratio_bar(0.0, width=5) == "----- 0%"
    assert format_ratio_bar(0.42, width=10) == "####------ 42%"
    assert format_ratio_bar(2.0, width=5) == "##### 100%"


def test_format_timestamp_handles_empty_values():
    assert format_timestamp(None) == "-"
    assert format_timestamp("2026-06-19T04:00:00Z") == "2026-06-19 04:00"
    assert format_timestamp("2026-06-10T20:42:51.498025999Z") == "2026-06-10 20:42"
