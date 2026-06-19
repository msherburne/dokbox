import os
import shutil
import subprocess
import tarfile
import zipfile
from pathlib import Path

try:
    import tomllib
except ModuleNotFoundError:  # pragma: no cover - Python < 3.11
    import tomli as tomllib


def _read_wheel_entry_points(archive: zipfile.ZipFile) -> str:
    entry_points_path = next(
        name
        for name in archive.namelist()
        if name.endswith(".dist-info/entry_points.txt")
    )
    return archive.read(entry_points_path).decode()


def _project_root() -> Path:
    return Path(__file__).resolve().parents[1]


def _run_common_helper(
    script: str,
    *,
    cwd: Path | None = None,
    env: dict[str, str] | None = None,
) -> subprocess.CompletedProcess[str]:
    project_root = _project_root()
    helper = project_root / "packaging" / "common.sh"
    command = f'source "{helper}"\n{script}'
    return subprocess.run(
        ["bash", "-c", command],
        check=True,
        cwd=cwd or project_root,
        capture_output=True,
        env=env,
        text=True,
    )


def test_pyproject_declares_build_backend_and_console_script():
    pyproject = _project_root() / "pyproject.toml"
    data = tomllib.loads(pyproject.read_text())

    build_system = data["build-system"]
    dev_dependencies = data["dependency-groups"]["dev"]

    assert build_system["build-backend"] == "hatchling.build"
    assert any(
        requirement.startswith("hatchling") for requirement in build_system["requires"]
    )
    assert data["project"]["scripts"]["dokbox"] == "dokbox.main:main"
    assert any(
        requirement.startswith("nuitka>=2.7.10") for requirement in dev_dependencies
    )
    assert any(requirement.startswith("tomli") for requirement in dev_dependencies)


def test_packaging_docs_cover_python_and_linux_build_commands():
    project_root = _project_root()
    root_readme = (project_root / "README.md").read_text()
    packaging_readme = (project_root / "packaging" / "README.md").read_text()

    assert "## Packaging" in root_readme
    assert "uv build" in root_readme
    assert "./packaging/build-linux.sh" in root_readme
    assert "Nuitka" in packaging_readme
    assert "uv build" in packaging_readme
    assert "./packaging/build-linux.sh" in packaging_readme
    assert "dokbox.main" in packaging_readme


def test_linux_build_helper_uses_nuitka_against_package_entrypoint():
    project_root = _project_root()
    helper = project_root / "packaging" / "build-linux.sh"
    helper_source = helper.read_text()

    assert helper.exists()
    assert helper.stat().st_mode & 0o111
    assert helper_source.startswith("#!/usr/bin/env bash")
    assert "set -euo pipefail" in helper_source
    assert "uv run python -m nuitka" in helper_source
    assert "--standalone" in helper_source
    assert "--output-dir=dist" in helper_source or "--output-dir dist" in helper_source
    assert "src/main.py" in helper_source
    assert "src/dokbox/main.py" not in helper_source
    assert 'DIST_DIR="dist"' in helper_source or "dist/" in helper_source


def test_release_workflow_runs_packaging_pipeline_and_uploads_dist_artifacts():
    workflow = (_project_root() / ".github" / "workflows" / "release.yml").read_text()

    assert "workflow_dispatch:" in workflow
    assert "uv sync --dev" in workflow
    assert "uv run pytest -v" in workflow
    assert "uv build" in workflow
    assert "./packaging/build-linux.sh" in workflow
    assert "uses: actions/upload-artifact@v4" in workflow
    assert "path: dist/*" in workflow
    assert "uv run semantic-release version" in workflow


def test_sdist_excludes_local_cache_artifacts(tmp_path):
    project_root = _project_root()

    subprocess.run(
        [
            "uv",
            "build",
            "--sdist",
            "--out-dir",
            str(tmp_path),
            "--no-create-gitignore",
        ],
        check=True,
        cwd=project_root,
    )

    sdist_path = next(tmp_path.glob("dokbox-*.tar.gz"))

    with tarfile.open(sdist_path, "r:gz") as archive:
        names = archive.getnames()

    assert not any("/.uv-cache/" in name for name in names)
    assert not any("/.venv/" in name for name in names)
    assert not any("/dokbox.egg-info/" in name for name in names)
    assert any(name.endswith("/pyproject.toml") for name in names)
    assert any(name.endswith("/README.md") for name in names)
    assert any(name.endswith("/src/dokbox/main.py") for name in names)
    assert any(name.endswith("/src/dokbox/ui/screens.py") for name in names)
    assert any(
        name.endswith("/src/dokbox/services/docker_service.py") for name in names
    )


def test_wheel_excludes_local_cache_and_metadata_artifacts(tmp_path):
    project_root = _project_root()

    subprocess.run(
        [
            "uv",
            "build",
            "--wheel",
            "--out-dir",
            str(tmp_path),
            "--no-create-gitignore",
        ],
        check=True,
        cwd=project_root,
    )

    wheel_path = next(tmp_path.glob("dokbox-*.whl"))

    with zipfile.ZipFile(wheel_path) as archive:
        names = archive.namelist()
        entry_points = _read_wheel_entry_points(archive)

    assert not any(".egg-info/" in name for name in names)
    assert not any("/.uv-cache/" in name for name in names)
    assert not any("/.venv/" in name for name in names)
    assert not any("/__pycache__/" in name for name in names)
    assert not any(name.endswith(".pyc") for name in names)
    assert "dokbox/main.py" in names
    assert "dokbox/ui/screens.py" in names
    assert "dokbox/services/docker_service.py" in names
    assert "main.py" not in names
    assert "[console_scripts]" in entry_points
    assert "dokbox = dokbox.main:main" in entry_points


def test_native_packaging_common_helper_defines_install_layout():
    project_root = _project_root()
    common = project_root / "packaging" / "common.sh"

    assert common.exists()

    source = common.read_text()

    assert 'PACKAGE_NAME="dokbox"' in source
    assert 'INSTALL_ROOT="/opt/dokbox"' in source
    assert 'BIN_PATH="/usr/bin/dokbox"' in source
    assert "stage_standalone_payload" in source
    assert "require_tool" in source


def test_common_helper_resolves_version_from_repo_root_without_uv_environment(tmp_path):
    pyproject = _project_root() / "pyproject.toml"
    expected_version = tomllib.loads(pyproject.read_text())["project"]["version"]
    outside_repo = tmp_path / "outside"
    outside_repo.mkdir()
    env = {
        "HOME": os.environ.get("HOME", str(tmp_path)),
        "PATH": os.environ["PATH"],
        "TERM": os.environ.get("TERM", "dumb"),
    }

    result = _run_common_helper(
        'printf "%s" "$PACKAGE_VERSION"',
        cwd=outside_repo,
        env=env,
    )

    assert result.stdout == expected_version


def test_common_helper_exposes_shared_metadata_through_bash():
    result = _run_common_helper(
        """
        printf '%s=%s\n' \
          PACKAGE_NAME "$PACKAGE_NAME" \
          PACKAGE_RELEASE "$PACKAGE_RELEASE" \
          HOST_ARCH "$HOST_ARCH" \
          PACKAGE_ARCH "$PACKAGE_ARCH" \
          STANDALONE_ARCH "$STANDALONE_ARCH" \
          STANDALONE_SUFFIX "$STANDALONE_SUFFIX" \
          STANDALONE_BASENAME "$STANDALONE_BASENAME" \
          STANDALONE_DIR "$STANDALONE_DIR" \
          STANDALONE_ARCHIVE "$STANDALONE_ARCHIVE" \
          PACKAGE_SUMMARY "$PACKAGE_SUMMARY" \
          PACKAGE_DESCRIPTION "$PACKAGE_DESCRIPTION" \
          PACKAGE_LICENSE "$PACKAGE_LICENSE" \
          PACKAGE_URL "$PACKAGE_URL" \
          PACKAGE_HOMEPAGE "$PACKAGE_HOMEPAGE" \
          PACKAGE_MAINTAINER "$PACKAGE_MAINTAINER" \
          PACKAGE_VENDOR "$PACKAGE_VENDOR" \
          PACKAGE_RUNTIME_DEPENDS "$PACKAGE_RUNTIME_DEPENDS"
        """
    )

    metadata = dict(line.split("=", 1) for line in result.stdout.strip().splitlines())

    assert metadata["PACKAGE_NAME"] == "dokbox"
    assert metadata["PACKAGE_RELEASE"] == "1"
    assert metadata["HOST_ARCH"]
    assert metadata["PACKAGE_ARCH"] == metadata["STANDALONE_ARCH"]
    assert metadata["STANDALONE_ARCH"]
    assert metadata["STANDALONE_SUFFIX"] == f"linux-{metadata['STANDALONE_ARCH']}"
    assert metadata["STANDALONE_BASENAME"] == f"dokbox-{metadata['STANDALONE_SUFFIX']}"
    assert metadata["STANDALONE_DIR"].endswith(
        f"/{metadata['STANDALONE_BASENAME']}.dist"
    )
    assert metadata["STANDALONE_ARCHIVE"].endswith(
        f"/{metadata['STANDALONE_BASENAME']}.tar.gz"
    )
    assert metadata["PACKAGE_SUMMARY"]
    assert metadata["PACKAGE_DESCRIPTION"]
    assert metadata["PACKAGE_LICENSE"]
    assert metadata["PACKAGE_URL"]
    assert metadata["PACKAGE_HOMEPAGE"] == metadata["PACKAGE_URL"]
    assert metadata["PACKAGE_MAINTAINER"]
    assert metadata["PACKAGE_VENDOR"]
    assert metadata["PACKAGE_RUNTIME_DEPENDS"] == ""


def test_stage_standalone_payload_creates_expected_layout(tmp_path):
    standalone_dir = tmp_path / "fake-standalone"
    stage_root = tmp_path / "stage"
    install_root = stage_root / "opt" / "dokbox"
    bin_link = stage_root / "usr" / "bin" / "dokbox"

    standalone_dir.mkdir()
    (standalone_dir / "dokbox").write_text("#!/bin/sh\nexit 0\n")
    (standalone_dir / "README.txt").write_text("standalone payload")

    _run_common_helper(
        f'''
        STANDALONE_DIR="{standalone_dir}"
        stage_standalone_payload "{stage_root}"
        ''',
        cwd=tmp_path,
    )

    assert install_root.is_dir()
    assert (install_root / "dokbox").read_text() == "#!/bin/sh\nexit 0\n"
    assert (install_root / "README.txt").read_text() == "standalone payload"
    assert bin_link.is_symlink()
    assert bin_link.readlink() == Path("/opt/dokbox/dokbox")


def test_debian_build_helper_stages_payload_and_invokes_dpkg_deb(tmp_path):
    project_root = _project_root()
    source_common = project_root / "packaging" / "common.sh"
    source_script = project_root / "packaging" / "build-deb.sh"

    workspace = tmp_path / "workspace"
    packaging_dir = workspace / "packaging"
    dist_dir = workspace / "dist"
    packaging_dir.mkdir(parents=True)
    dist_dir.mkdir()

    shutil.copy2(source_common, packaging_dir / "common.sh")
    shutil.copy2(source_script, packaging_dir / "build-deb.sh")
    (packaging_dir / "build-linux.sh").write_text(
        "#!/usr/bin/env bash\n"
        "printf 'unexpected build-linux invocation\\n' >&2\n"
        "exit 99\n",
        encoding="utf-8",
    )
    (packaging_dir / "build-linux.sh").chmod(0o755)

    (workspace / "pyproject.toml").write_text(
        '[project]\nname = "dokbox"\nversion = "9.8.7"\n',
        encoding="utf-8",
    )

    standalone_dir = dist_dir / "dokbox-linux-x86_64.dist"
    standalone_dir.mkdir()
    (standalone_dir / "dokbox").write_text("#!/bin/sh\nexit 0\n", encoding="utf-8")
    (standalone_dir / "asset.txt").write_text("payload asset\n", encoding="utf-8")

    fake_bin = tmp_path / "fake-bin"
    fake_bin.mkdir()
    dpkg_deb_log = tmp_path / "dpkg-deb.log"
    (fake_bin / "dpkg-deb").write_text(
        "#!/usr/bin/env bash\n"
        "set -euo pipefail\n"
        'printf \'%s\\n\' "$@" > "$DPKG_DEB_LOG"\n'
        'stage_root="${@: -2:1}"\n'
        'output_path="${@: -1}"\n'
        'mkdir -p "$(dirname "$output_path")"\n'
        'printf "fake deb\\n" > "$output_path"\n'
        'test -f "$stage_root/DEBIAN/control"\n'
        'test -f "$stage_root/opt/dokbox/dokbox"\n'
        'test -f "$stage_root/opt/dokbox/asset.txt"\n'
        'test -L "$stage_root/usr/bin/dokbox"\n',
        encoding="utf-8",
    )
    (fake_bin / "dpkg-deb").chmod(0o755)

    env = {
        "DPKG_DEB_LOG": str(dpkg_deb_log),
        "HOME": os.environ.get("HOME", str(tmp_path)),
        "PATH": f"{fake_bin}:{os.environ['PATH']}",
        "TERM": os.environ.get("TERM", "dumb"),
    }

    result = subprocess.run(
        ["bash", str(packaging_dir / "build-deb.sh")],
        cwd=workspace,
        check=True,
        capture_output=True,
        env=env,
        text=True,
    )

    deb_path = dist_dir / "dokbox_9.8.7-1_amd64.deb"
    control = (dist_dir / "deb" / "DEBIAN" / "control").read_text(encoding="utf-8")
    bin_link = dist_dir / "deb" / "usr" / "bin" / "dokbox"
    dpkg_deb_args = dpkg_deb_log.read_text(encoding="utf-8").splitlines()

    assert source_script.exists()
    assert source_script.stat().st_mode & 0o111
    assert deb_path.exists()
    assert "Built" in result.stdout
    assert str(deb_path) in result.stdout
    assert dpkg_deb_args[:2] == ["--build", str(dist_dir / "deb")]
    assert dpkg_deb_args[-1] == str(deb_path)
    assert "Package: dokbox" in control
    assert "Version: 9.8.7-1" in control
    assert "Architecture: amd64" in control
    assert "Depends:" not in control
    assert "\n\nHomepage:" not in control
    assert "Maintainer: dokbox maintainers" in control
    assert "Homepage: https://github.com/msherburne/dokbox" in control
    assert "Description: Terminal UI for Docker resource management" in control
    assert bin_link.is_symlink()
    assert bin_link.readlink() == Path("/opt/dokbox/dokbox")
