# Recipe: Release A Library Version

Use this to publish a reusable Python distribution; deployable service releases use the release checklist.

## Files To Touch

- `[project].version` in `pyproject.toml`, `uv.lock`, and `CHANGELOG.md`
- package exports and `src/<package>/py.typed`
- protected `v<version>` tag and release workflow

## Steps

1. Choose the version from public API/behavior compatibility; breaking change means major, additive feature minor, fix patch.
2. Run `uv version <version>` and move `Unreleased` entries to `[<version>] - YYYY-MM-DD`.
3. Run `make verify` and `uv build --no-sources`; inspect sdist and wheel for metadata, license, intended modules, and `py.typed`.
4. Install the wheel in isolation and import the package/run declared scripts without the checkout.
5. Create the matching protected/signed `v<version>` tag; the protected workflow publishes once with PyPI Trusted Publishing via `uv publish`.

## Invariants To Preserve

- PEP 621 static version, changelog, tag, and artifact metadata agree.
- The wheel is built from the sdist path and contains no tests, caches, secrets, or local source overrides.
- Supported Python floor is exercised; typed libraries ship `py.typed` and expose intentional public symbols only.
- Publishing never occurs from a PR or with a long-lived token when Trusted Publishing is available.

## Proof

```bash
make verify
uv build --no-sources
uv run --with ./dist/<wheel> --no-project -- python -c "import <package>"
python -m zipfile -l ./dist/<wheel>
git cat-file -t v<version>
```

Verify the published artifact from the registry after release. Governing doc: [CI and release](../operations/ci-and-release.md).
