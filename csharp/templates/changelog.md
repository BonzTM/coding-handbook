<!--
Template: CHANGELOG.md for a repo built on the C# handbook.
Copy to the repo root as CHANGELOG.md and replace every <PLACEHOLDER>.
Format: Keep a Changelog 1.1.0 (https://keepachangelog.com/en/1.1.0/).
Versioning: Semantic Versioning 2.0.0 (https://semver.org/).

Rules:
- Newest version first. Keep an [Unreleased] section at the top.
- On release, rename [Unreleased] to the new version + date (YYYY-MM-DD),
  then add a fresh empty [Unreleased] above it.
- Use only these section names, omit the ones with no entries:
    Added       — new features
    Changed     — changes in existing functionality
    Deprecated  — soon-to-be removed features
    Removed     — now removed features
    Fixed       — bug fixes
    Security    — vulnerability fixes
- Release tags are v<MAJOR>.<MINOR>.<PATCH> (the "v" prefix is required; the
  release workflow triggers on it). The headings below drop the "v" per Keep a
  Changelog, and so does the NuGet PackageVersion for libraries — tag v1.2.3
  <-> heading [1.2.3] <-> package 1.2.3.
- Every operator-visible change (config keys, migrations, ports, contract
  changes) MUST have an entry here. See <HANDBOOK_URL>/operations/ci-and-release.md.
-->

# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- <NEW_FEATURE_OR_ENDPOINT>

### Changed

- <BEHAVIOR_OR_DEFAULT_THAT_CHANGED>

### Deprecated

- <FEATURE_MARKED_FOR_REMOVAL_AND_ITS_REPLACEMENT>

### Removed

- <FEATURE_OR_CONTRACT_REMOVED>

### Fixed

- <USER_VISIBLE_BUG_FIX>

### Security

- <VULNERABILITY_FIX_OR_DEPENDENCY_BUMP>

## [1.0.0] - <YYYY-MM-DD>

### Added

- <INITIAL_RELEASE_HIGHLIGHTS>

<!--
Comparison links (optional but recommended). Update the tags on each release.

[Unreleased]: <REPO_URL>/compare/v1.0.0...HEAD
[1.0.0]: <REPO_URL>/releases/tag/v1.0.0
-->
