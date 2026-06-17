<!--
Destination: SECURITY.md at repo root, or .github/SECURITY.md (GitHub renders either as the
"Security policy" tab; .github/ keeps the root clean).

Contract:
- Every externally-facing service and every published library ships this file.
  Internal-only services may use a trimmed version but MUST still name an owner and a private
  report path (see golang/operations/security.md ### Vulnerability Disclosure).
- It exists so a reporter never has to open a PUBLIC issue for an unpatched vulnerability.
- Keep the supported-versions table current with what you actually patch; a stale table is a defect.
- Replace every <PLACEHOLDER>. Delete rows/sections that genuinely do not apply.
-->

# Security Policy

We take the security of <PROJECT_NAME> seriously. This document explains which versions
receive security fixes and how to report a vulnerability privately.

## Supported Versions

Security fixes are backported only to the versions below. Versions marked unsupported get no
fixes — upgrade to a supported release. Versions follow `vMAJOR.MINOR.PATCH`.

| Version | Supported |
|---|---|
| <v1.4.x> | :white_check_mark: |
| <v1.3.x> | :white_check_mark: |
| <v1.2.x and older> | :x: |

## Reporting a Vulnerability

**Do not open a public issue, pull request, or discussion for an unpatched vulnerability.**
Public disclosure before a fix ships puts every user at risk.

Report privately through one of:

- **GitHub private security advisory** (preferred): <REPO_URL>/security/advisories/new
- **Email:** <security@ORG_DOMAIN> (<OPTIONAL: PGP key fingerprint / link>)

Please include enough for us to reproduce and assess impact:

- affected version(s), and the commit or release tag if known;
- a description of the vulnerability and its impact (what an attacker can do);
- step-by-step reproduction, proof-of-concept, or a failing test;
- any known mitigations or workarounds;
- how you would like to be credited (or that you prefer to stay anonymous).

## Response & Disclosure

- **Acknowledgement:** we aim to acknowledge your report within <2 business days>.
- **Triage:** we aim to confirm the vulnerability and assess severity within <5 business days>,
  and will keep you updated at least <weekly> until it is resolved.
- **Coordinated disclosure:** we follow a coordinated-disclosure model. We ask that you keep the
  report private until a fix is released or <90 days> have passed, whichever comes first. We will
  agree an embargo and disclosure date with you and publish a GitHub Security Advisory (with CVE
  where applicable) when the fix ships.
- **Fix & release:** we will develop the fix privately, ship it in a supported release, and
  document it in the advisory and release notes.
- **Credit:** with your consent, we credit reporters in the advisory and release notes. We do not
  currently run a paid bounty program. <OPTIONAL: bounty / safe-harbor statement.>

Thank you for helping keep <PROJECT_NAME> and its users safe.
