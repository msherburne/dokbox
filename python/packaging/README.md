# Packaging

Build the standard Python artifacts with:

```bash
uv build
```

Build the Linux standalone bundle with Nuitka using the packaged entrypoint:

```bash
./packaging/build-linux.sh
```

The helper:

- runs `uv run python -m nuitka`
- builds the `dokbox.main` entrypoint from `src/dokbox/main.py`
- targets `src/dokbox/main.py` with `PYTHONPATH=src` so the new `dokbox`
  package layout resolves correctly
- writes `dist/dokbox-linux-x86_64.dist/` and `dist/dokbox-linux-x86_64.tar.gz`

Run `uv sync --dev` first so Nuitka and the project dependencies are available
in the local environment. On Linux, install `patchelf` as well, because
Nuitka standalone builds require it.
