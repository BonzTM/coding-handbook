# CODEOWNERS template -> .github/CODEOWNERS
#
# Placement: this file lives at .github/CODEOWNERS (GitHub also accepts repo-root
# or docs/ CODEOWNERS; pick ONE). It auto-requests review from the listed owners
# when matching paths change, and pairs with the repo AGENTS.md Change Routing
# table: routing says which doc to read, CODEOWNERS says who must approve.
# Keep the two in sync — same package areas, same owners.
#
# Rules: last matching pattern wins, so order general -> specific. Owners are
# GitHub @users or @org/teams and must have write access. Replace every <owner>.

# Fallback owner for anything not matched below.
*                       <@org/team-platform>

# Process wiring and shutdown — no business logic lives here.
/cmd/                   <@org/team-platform>

# Domain logic and the interfaces consumed from the outside.
/internal/core/         <@org/team-domain>

# Transport adapters only.
/internal/api/          <@org/team-api>
/internal/api/grpc/     <@org/team-api>

# Queries, migrations, transactions, storage mapping.
/internal/db/           <@org/team-data>

# Loading, defaults, validation, fail-fast startup. Pairs with .env.example + README config table.
/internal/config/       <@org/team-platform>

# Logging, metrics, tracing, health primitives.
/internal/telemetry/    <@org/team-observability>

# Published wire contracts (.proto, OpenAPI). Contract changes need extra scrutiny.
/api/                   <@org/team-api>

# Build, CI, release automation.
/.github/               <@org/team-platform>
/Makefile               <@org/team-platform>

# Keep the contributor contract owned so its invariants do not drift.
/AGENTS.md              <@org/team-platform>
/README.md              <@org/team-platform>
