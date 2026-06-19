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
