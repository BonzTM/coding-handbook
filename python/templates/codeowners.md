# Destination: .github/CODEOWNERS. Replace every <owner> and <app>.
# Last matching pattern wins; order general to specific.
*                           <@org/team-platform>
/src/<app>/core/            <@org/team-domain>
/src/<app>/api/             <@org/team-api>
/src/<app>/clients/         <@org/team-integrations>
/src/<app>/db/              <@org/team-data>
/src/<app>/workers/         <@org/team-messaging>
/src/<app>/telemetry/       <@org/team-observability>
/alembic/                   <@org/team-data>
/api/                       <@org/team-api>
/.github/                   <@org/team-platform>
/Makefile                   <@org/team-platform>
/pyproject.toml             <@org/team-platform>
/uv.lock                    <@org/team-platform>
/AGENTS.md                  <@org/team-platform>
/README.md                  <@org/team-platform>
