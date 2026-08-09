# Recipe: Bump A Dependency

Use this for one direct or security-driven dependency upgrade.

## Files To Touch

- `package.json` and `package-lock.json`
- callers and adapters affected by API or behavior changes
- tests for the dependency-owned boundary
- templates and docs when a handbook pin changes
- ADR, changelog, or security exception when required

## Steps

1. Record current and target versions, release type, driver, and supported Node/ESM requirements.
2. Read official release notes and advisories; identify breaking and behavioral changes.
3. Re-run framework approval questions for a major or materially expanded dependency.
4. Install the exact target through npm:

   ```bash
   npm install --save-exact <package>@<version>
   ```

5. Review both manifest and lockfile; explain every material transitive or install-script change.
6. Update the narrow adapter and negative-path tests.
7. Run `npm ls <package>` and the dependency-specific smoke test.
8. Run audit and record any reachable exception with owner and expiry.
9. Run the clean canonical gate on Node 24.

## Invariants To Preserve

- One npm lockfile remains authoritative and `npm ci` does not rewrite it.
- No unrelated dependency, runtime, module, or package-manager migration rides along.
- New permissions, binaries, scripts, licenses, and transitive risks are reviewed.
- Direct tool versions remain exact pins.
- Major behavior or architecture changes have an accepted ADR.
- Security exceptions are scoped, owned, compensated, and time-bounded.

## Proof

- `git diff -- package.json package-lock.json` is reviewed.
- `npm ci` succeeds from a clean checkout on Node 24.
- Targeted compatibility and negative-path tests pass.
- `npm audit --audit-level=high` passes or links an accepted exception.
- `npm run verify` is green.
