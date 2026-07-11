# Recipe: Add Database Feature

Use this when a feature changes schema, queries, or transaction behavior.

Governing doc: [`csharp/services/database.md`](../services/database.md).

## Files To Touch

- the entity type and its `IEntityTypeConfiguration<T>` under `src/Orders.Infrastructure/Data/Configurations`
- the `DbContext` (`DbSet` + `ApplyConfiguration`) and a new EF Core migration under `src/Orders.Infrastructure/Data/Migrations` (see [add-migration.md](add-migration.md))
- the repository implementing the port under `src/Orders.Infrastructure/Data`, and the port in `src/Orders.Core/Ports` if the seam changes
- `src/Orders.Core/...` if domain behavior changes
- Testcontainers integration tests in `tests/Orders.IntegrationTests`

## Steps

1. Model the change: the domain type lives in `Orders.Core`; its mapping (`IEntityTypeConfiguration<T>` — table, keys, indexes, column types, constraints) lives in `Orders.Infrastructure/Data/Configurations`. Core never references EF Core.
2. Add the migration (`dotnet ef migrations add <Name> --project src/Orders.Infrastructure --startup-project src/Orders.Api`), review the generated code and SQL, and decide whether it is deploy-safe with mixed-version application instances (expand/contract: additive now, destructive in a later release).
3. Update the repository behind the Core port. Queries are LINQ (parameterized by construction); raw SQL only via the interpolated `FromSql`/`SqlQuery` overloads, never `FromSqlRaw` with concatenated input. Storage-specific mapping stays out of Core.
4. Update the Core service to call the port, not the `DbContext`. The coordinating layer owns the unit of work: one `SaveChanges` per use case, or an explicit transaction at the coordinating seam when a use case spans multiple aggregates — repositories never commit on their own.
5. Add Testcontainers.PostgreSql integration tests that apply the migrations to a clean database (`Database.MigrateAsync()` in the test fixture only) and exercise the real storage path, behind the `-Integration` switch.

## Invariants To Preserve

- EF Core, SQL, and provider types stay in `Orders.Infrastructure`; Core sees only the port
- transactions are owned by the coordinating layer, not by repositories
- queries remain parameterized; no string-built SQL
- migrations and application code can coexist safely during rollout
- migrations run only via the explicit migration step (`--migrate` flag or init job) — never auto-migrate on normal startup

## Proof

- migration apply test on a clean Testcontainers database
- integration test for the changed behavior against the real database path
- model-drift check: `dotnet ef migrations has-pending-model-changes --project src/Orders.Infrastructure --startup-project src/Orders.Api` reports none
- rollback or compatibility review for destructive schema changes
- run `pwsh ./verify.ps1` (integration suite behind `-Integration`)
