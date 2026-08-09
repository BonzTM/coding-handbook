# Recipe: Release A Library Version

Use this when publishing a TypeScript package version.

## Files To Touch

- `package.json`, package exports, and lockfile
- public declarations and documentation
- `CHANGELOG.md` and release notes
- clean consumer fixture
- release workflow and provenance metadata

## Steps

1. Classify changes across runtime exports, declarations, errors, side effects, and platform support.
2. Select the semantic version; any incompatible public behavior requires a major.
3. Move relevant changelog entries from `Unreleased` to the release version and date.
4. Confirm every documented subpath appears in `exports` and internal paths do not.
5. Run the canonical gate from a clean Node 24 install.
6. Build once, then create and inspect the tarball:

   ```bash
   npm pack --dry-run
   npm pack
   ```

7. Install the tarball in a clean consumer and test ESM runtime plus declarations.
8. Publish from an approved protected workflow using trusted or short-lived identity.
9. Tag the published commit and verify registry metadata, provenance, and consumer install.

## Invariants To Preserve

- The release artifact comes from the verified commit and is not rebuilt per environment.
- Public declarations expose no dependency-private or source-only paths.
- The packed file list contains no secrets, fixtures, local config, or unintended source maps.
- Version, tag, changelog, tarball, and registry metadata agree.
- Failed or partial publication has a documented recovery path.
- Credentials are unavailable to untrusted pull-request code.

## Proof

- `npm ci && npm run verify` is green on the release commit.
- `npm pack --dry-run` output is reviewed.
- A clean consumer imports every documented path and typechecks.
- The published digest/version maps to source revision and provenance.
- Release notes state breaking, operator-visible, and security changes.
