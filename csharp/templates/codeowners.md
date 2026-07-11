# CODEOWNERS template -> .github/CODEOWNERS
#
# Placement: this file lives at .github/CODEOWNERS (GitHub also accepts repo-root
# or docs/ CODEOWNERS; pick ONE). It auto-requests review from the listed owners
# when matching paths change, and pairs with the repo AGENTS.md Change Routing
# table: routing says which doc to read, CODEOWNERS says who must approve.
# Keep the two in sync — same project areas, same owners.
#
# Rules: last matching pattern wins, so order general -> specific. Owners are
# GitHub @users or @org/teams and must have write access. Replace every <owner>
# and every <App> with the real project names.

# Fallback owner for anything not matched below.
*                                   <@org/team-platform>

# Thin host: composition root, endpoint groups, telemetry wiring — no business logic.
/src/<App>.Api/                     <@org/team-api>

# Domain types, ports, business rules.
/src/<App>.Core/                    <@org/team-domain>

# EF Core, migrations, repositories, outbound clients.
/src/<App>.Infrastructure/          <@org/team-data>

# Logging, metrics, tracing, health wiring.
/src/<App>.Api/Telemetry/           <@org/team-observability>

# Published wire contracts (.proto, OpenAPI). Contract changes need extra scrutiny.
/api/                               <@org/team-api>

# Build, CI, release automation, and the single verification gate.
/.github/                           <@org/team-platform>
/verify.ps1                         <@org/team-platform>
/Makefile                           <@org/team-platform>

# Toolchain and dependency pins: SDK version, build defaults, central package
# versions, feeds. Changes here affect every project in the solution.
/global.json                        <@org/team-platform>
/Directory.Build.props              <@org/team-platform>
/Directory.Packages.props           <@org/team-platform>
/nuget.config                       <@org/team-platform>

# Keep the contributor contract owned so its invariants do not drift.
/AGENTS.md                          <@org/team-platform>
/README.md                          <@org/team-platform>
