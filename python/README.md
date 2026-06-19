## Release workflow

This project uses conventional commits.

- `feat:` releases a minor version
- `fix:` and `perf:` release a patch version
- `BREAKING CHANGE:` or `!` triggers a major version

Install hooks with:

```bash
uv run pre-commit install --hook-type pre-commit --hook-type commit-msg
```

You can also create commits interactively with:

```bash
uv run cz commit
```

When pull requests merge to `main`, GitHub Actions runs semantic release to update `pyproject.toml`, refresh `CHANGELOG.md`, create a tag, and publish a GitHub release.

## Testing

Run the test suite with:

```bash
uv run pytest
```

## Packaging

Build the wheel and sdist with:

```bash
uv build
```

Build the Linux standalone artifact with:

```bash
./packaging/build-linux.sh
```

The helper expects a Linux build host with the dev environment installed via
`uv sync --dev`, plus `patchelf` on Linux for Nuitka standalone builds, and
writes both the standalone folder and a `.tar.gz` archive to `dist/`.
