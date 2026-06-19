from datetime import datetime


def format_bytes(value: int | float | None) -> str:
    if value is None:
        return "-"
    amount = float(value)
    units = ["B", "KiB", "MiB", "GiB", "TiB"]
    unit = units[0]
    for unit in units:
        if abs(amount) < 1024 or unit == units[-1]:
            break
        amount /= 1024
    if unit == "B":
        return f"{int(amount)} {unit}"
    return f"{amount:.1f} {unit}"


def format_ratio_bar(ratio: float | None, width: int = 10) -> str:
    if ratio is None:
        return f"{'-' * width} n/a"
    clamped = max(0.0, min(ratio, 1.0))
    filled = round(clamped * width)
    empty = width - filled
    return f"{'#' * filled}{'-' * empty} {clamped:.0%}"


def format_timestamp(value: str | None) -> str:
    if not value:
        return "-"
    normalized = value.replace("Z", "+00:00")
    parsed = datetime.fromisoformat(normalized)
    return parsed.strftime("%Y-%m-%d %H:%M")
