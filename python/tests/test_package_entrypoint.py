import ast
from importlib import import_module
from importlib.util import module_from_spec, spec_from_file_location
from pathlib import Path
import sys

import pytest
from dokbox.models.config import DokboxConfig

package_main = import_module("dokbox.main")

OWNED_TEST_FILES = [
    "tests/test_config.py",
    "tests/test_docker_service.py",
    "tests/test_formatting.py",
    "tests/test_resource_view_models.py",
    "tests/test_setup.py",
    "tests/test_ui_smoke.py",
    "tests/test_package_entrypoint.py",
]


def test_package_init_does_not_shadow_main_submodule():
    sys.modules.pop("dokbox.main", None)
    sys.modules.pop("dokbox", None)

    package = import_module("dokbox")

    with pytest.raises(AttributeError):
        package.main

    package_main = import_module("dokbox.main")

    assert package.main is package_main


def test_package_main_loads_config_and_runs_app(monkeypatch):
    config = DokboxConfig(docker_host="unix:///tmp/docker.sock")
    calls: list[tuple[str, object]] = []

    class FakeDokbox:
        def __init__(self, received_config):
            calls.append(("init", received_config))

        def run(self):
            calls.append(("run", None))

    monkeypatch.setattr(package_main, "load_or_setup_config", lambda: config)
    monkeypatch.setattr(package_main, "Dokbox", FakeDokbox)

    package_main.main()

    assert calls == [("init", config), ("run", None)]


def test_package_dokbox_constructs_default_docker_service(monkeypatch):
    config = DokboxConfig(docker_host="unix:///tmp/docker.sock")
    docker_client = object()
    load_calls: list[str] = []
    service_clients: list[object] = []

    monkeypatch.setattr(
        package_main,
        "load_docker_client",
        lambda docker_host: load_calls.append(docker_host) or docker_client,
    )

    class FakeDockerService:
        def __init__(self, client):
            service_clients.append(client)

    monkeypatch.setattr(package_main, "DockerService", FakeDockerService)

    app = package_main.Dokbox(config)

    assert app.docker_client is docker_client
    assert isinstance(app.docker_service, FakeDockerService)
    assert load_calls == ["unix:///tmp/docker.sock"]
    assert service_clients == [docker_client]


def test_legacy_top_level_imports_are_not_used_in_tests():
    legacy_roots = {
        "main",
        "defaults",
        "models",
        "services",
        "ui",
        "utils",
        "view_models",
    }
    offenders: list[str] = []
    repo_root = Path(__file__).resolve().parents[1]

    for relative_path in OWNED_TEST_FILES:
        test_file = repo_root / relative_path
        tree = ast.parse(test_file.read_text(), filename=relative_path)

        for node in ast.walk(tree):
            if isinstance(node, ast.Import):
                for alias in node.names:
                    root = alias.name.split(".", 1)[0]
                    if root in legacy_roots:
                        offenders.append(
                            f"{relative_path}:{node.lineno}: import {alias.name}"
                        )
            elif isinstance(node, ast.ImportFrom):
                if node.module is None:
                    continue
                root = node.module.split(".", 1)[0]
                if root in legacy_roots:
                    offenders.append(
                        f"{relative_path}:{node.lineno}: from {node.module} import ..."
                    )
            elif (
                isinstance(node, ast.Call)
                and isinstance(node.func, ast.Attribute)
                and node.func.attr == "insert"
                and isinstance(node.func.value, ast.Attribute)
                and node.func.value.attr == "path"
                and isinstance(node.func.value.value, ast.Name)
                and node.func.value.value.id == "sys"
            ):
                offenders.append(f"{relative_path}:{node.lineno}: sys.path.insert(...)")
            elif (
                isinstance(node, ast.Call)
                and isinstance(node.func, ast.Name)
                and node.func.id == "import_module"
                and node.args
                and isinstance(node.args[0], ast.Constant)
                and node.args[0].value == "main"
            ):
                offenders.append(
                    f'{relative_path}:{node.lineno}: import_module("main")'
                )

    assert offenders == []


def test_package_main_wrapper_reexports_package_entrypoint(monkeypatch):
    current_package_main = import_module("dokbox.main")
    wrapper_path = Path(__file__).resolve().parents[1] / "src" / "main.py"
    spec = spec_from_file_location("legacy_main_wrapper", wrapper_path)
    assert spec is not None
    assert spec.loader is not None
    wrapper_main = module_from_spec(spec)
    spec.loader.exec_module(wrapper_main)

    assert wrapper_main.Dokbox.__name__ == "Dokbox"
    assert wrapper_main.Dokbox.__module__ == "dokbox.main"
    assert wrapper_main.main.__name__ == "main"
    assert wrapper_main.main.__module__ == "dokbox.main"

    config = DokboxConfig(docker_host="unix:///tmp/docker.sock")
    calls: list[tuple[str, object]] = []

    class FakeDokbox:
        def __init__(self, received_config):
            calls.append(("init", received_config))

        def run(self):
            calls.append(("run", None))

    monkeypatch.setattr(current_package_main, "load_or_setup_config", lambda: config)
    monkeypatch.setattr(current_package_main, "Dokbox", FakeDokbox)

    wrapper_main.main()

    assert calls == [("init", config), ("run", None)]


def test_all_runtime_modules_import_from_dokbox_namespace():
    runtime_modules = [
        "dokbox.defaults.config",
        "dokbox.models.config",
        "dokbox.models.docker_resources",
        "dokbox.services.docker_service",
        "dokbox.ui.screens",
        "dokbox.ui.widgets",
        "dokbox.utils.cli",
        "dokbox.utils.config",
        "dokbox.utils.docker",
        "dokbox.utils.setup",
        "dokbox.view_models.formatting",
        "dokbox.view_models.resources",
    ]

    for module_name in runtime_modules:
        module = import_module(module_name)
        module_path = Path(module.__file__).resolve()

        assert "/src/dokbox/" in module_path.as_posix(), (
            f"{module_name} resolved to {module_path}, expected a module under src/dokbox"
        )
