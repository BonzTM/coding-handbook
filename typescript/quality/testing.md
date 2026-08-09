# Testing

Testing strategy for TypeScript repositories that need trustworthy behavior rather than green checkmarks.

## Default Approach

Test observable contracts at the narrowest boundary that proves them, then add a small number of real integration paths.

### Test Taxonomy

| Test type | Use for | Default tools |
|---|---|---|
| unit | pure domain rules, parsers, mappers, retry decisions | Jest 30 |
| transport | Fastify decoding, auth, status and problem mapping | `app.inject()` |
| component | accessible React behavior and user interactions | React Testing Library, user-event |
| integration | PostgreSQL, migrations, multi-module behavior | Testcontainers |
| contract | HTTP/event/library compatibility and golden payloads | Zod schemas, fixtures, packed consumers |
| external-boundary | browser HTTP behavior and backend client protocol | MSW; local HTTP server |
| end-to-end | a few critical deployed user or operator journeys | project-approved browser/deployment harness |

Keep unit and component tests fast and offline. Put real infrastructure behind an explicit integration script or flag and run it in CI. Do not call a mocked database test “integration.”

### Jest 30 And ESM

Jest 30 native ESM support remains experimental and requires Node's `--experimental-vm-modules`. The handbook default is therefore a pinned transform-based setup, not a claim that Jest runs TypeScript ESM transparently.

Pin Jest, its TypeScript transformer, and matching type packages in the lockfile. Keep one committed Jest configuration that transforms `.ts` and `.tsx` for the test runtime, maps source imports consistently, uses `jsdom` only for frontend projects, and clears or restores mocks between tests.

The backend baseline uses Jest 30 with `babel-jest`, `@babel/preset-env`, and `@babel/preset-typescript`. In an ESM package, commit `babel.config.cjs`:

```js
module.exports = {
  presets: [
    ["@babel/preset-env", { targets: { node: "24" }, modules: "commonjs" }],
    "@babel/preset-typescript",
  ],
};
```

Commit `jest.config.cjs` so the configuration itself has unambiguous CommonJS semantics:

```js
/** @type {import("jest").Config} */
module.exports = {
  testEnvironment: "node",
  testMatch: ["<rootDir>/src/**/*.test.ts"],
  transform: { "^.+\\.tsx?$": "babel-jest" },
  moduleNameMapper: { "^(\\.{1,2}/.*)\\.js$": "$1" },
  clearMocks: true,
  restoreMocks: true,
};
```

Run this transform-based configuration with plain `jest`; do not set `NODE_OPTIONS=--experimental-vm-modules`, and do not set `extensionsToTreatAsEsm`. Babel emits CommonJS for the Jest runtime and performs no type checking, so `tsc --noEmit` is a separate mandatory gate. The `.js` mapper lets NodeNext-authored relative imports resolve their `.ts` sources in Jest; the production-artifact smoke test proves real emitted resolution.

The frontend variant adds JSX transformation and jsdom. Add `@babel/preset-react` with `{ runtime: "automatic" }`, then use:

```js
module.exports = {
  ...require("./jest.config.cjs"),
  testEnvironment: "jsdom",
  testMatch: ["<rootDir>/src/**/*.test.{ts,tsx}"],
  setupFilesAfterEnv: ["<rootDir>/src/test/setup.ts"],
};
```

`src/test/setup.ts` imports `@testing-library/jest-dom`. Keep backend and frontend as separate Jest projects when both shapes share one repository.

Production remains ESM and is built by `tsc` for backends or Vite for frontends. The test transform is test infrastructure; it does not define production module behavior. Add a production-artifact smoke test so transform differences cannot conceal bad NodeNext imports or package exports.

If a repository chooses Jest's native ESM path, it must pin the exact working command and config, include `--experimental-vm-modules`, avoid unsupported mocking assumptions, and record why the experimental path is worth its maintenance cost. Vitest remains an ADR-governed alternative for Vite-heavy applications.

### Test Structure And Naming

Name tests by behavior and outcome: `rejects an expired idempotency key`, not `test case 3`. Use arrange/act/assert separation when it improves readability, without ritual comments.

Keep tests beside the module they exercise. Put cross-boundary and infrastructure suites in a named integration directory with a distinct Jest project or script. One command must select one test or suite for fast diagnosis.

Prefer focused cases over giant parameter tables. Use `test.each` only when setup, action, and assertion are genuinely identical across inputs.

### Test Doubles

Use small hand-written fakes at consumer-owned ports. A fake clock, in-memory implementation, controlled promise, or stub logger should model behavior, not a transcript of internal calls.

Mock only boundaries the test does not own. Do not mock the subject, Zod parsing, SQL semantics, or React hooks merely to make the test easy. Over-specified call order couples tests to refactoring and misses observable defects.

Reset spies and mocks after each test. Avoid `jest.mock` hoisting surprises by keeping module mocking rare and localized. Explicit dependency injection is the preferred seam.

### React Testing Library

Test the interface as a user perceives it. Query priority is:

1. `getByRole` with accessible name;
2. `getByLabelText` for form controls;
3. `getByPlaceholderText` only when placeholder is the actual accessible affordance;
4. `getByText` and `getByDisplayValue` for visible content;
5. semantic test IDs only when no accessible query can express the contract.

Use `userEvent.setup()` and await interactions. Prefer `findBy*` for an element that will appear and `waitFor` for an assertion that must eventually become true. Do not wrap every action in `waitFor` or use arbitrary delays.

Cover keyboard navigation, accessible name, focus, validation association, loading, empty, error, success, disabled, and retry states where applicable. Assert behavior rather than component internals, hook calls, snapshots, or CSS class implementation.

```tsx
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { SaveButton } from "./save-button.js";

it("submits through the accessible control", async () => {
  const user = userEvent.setup();
  const onSave = jest.fn<() => Promise<void>>().mockResolvedValue();
  render(<SaveButton onSave={onSave} />);

  await user.click(screen.getByRole("button", { name: "Save widget" }));

  expect(onSave).toHaveBeenCalledTimes(1);
});
```

### MSW For HTTP Fakes

Use MSW at the network boundary for frontend API tests. Handlers validate method, path, required headers, and request body before returning a response. Unhandled requests fail the test.

Define default happy-path handlers centrally and override narrowly per test. Reset overrides after each test and close the server after the suite. Cover latency through controlled timers or promises, HTTP errors, malformed success payloads, abort, retry, and network failure.

MSW does not prove the real service implements the same contract. Pair it with shared contract fixtures or consumer/provider contract tests where independent deployments make drift material.

Backend external HTTP clients use a local in-process HTTP server by default so URL resolution, headers, timeout, abort, size bounds, and parsing execute for real. Do not call public sandboxes from the deterministic suite.

```ts
import { afterAll, afterEach, beforeAll } from "@jest/globals";
import { http, HttpResponse } from "msw";
import { setupServer } from "msw/node";

const server = setupServer(
  http.get("http://localhost/api/widgets/:id", ({ params }) =>
    HttpResponse.json({ id: params.id, name: "Meter" }),
  ),
);

beforeAll(() => server.listen({ onUnhandledRequest: "error" }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());
```

### PostgreSQL Integration With Testcontainers

Database adapters and migrations run against a real, version-pinned PostgreSQL container. Mocked `pg` results cannot prove SQL, types, constraints, locks, transaction behavior, or migration compatibility.

Start the container once per suite or Jest worker according to measured isolation needs. Apply migrations through the same supported migration command used by deployment. Give each test a clean schema, database, or transaction strategy that cannot leak state between tests.

Cap startup and query timeouts. Always stop containers and close pools in `finally` or suite teardown. Integration CI needs bounded worker concurrency so parallel containers do not exhaust the runner.

Test empty-database migration, upgrade from the prior supported schema, constraints, row parsing, transaction rollback, concurrency conflicts, and representative queries. When SQL shape matters for performance, capture `EXPLAIN` evidence separately from ordinary correctness tests.

```ts
import { PostgreSqlContainer, type StartedPostgreSqlContainer } from "@testcontainers/postgresql";
import { Pool } from "pg";
import { runMigrations } from "../db/migrate.js";

let container: StartedPostgreSqlContainer;
let pool: Pool;

beforeAll(async () => {
  container = await new PostgreSqlContainer("postgres:17.6").start();
  pool = new Pool({ connectionString: container.getConnectionUri() });
  await runMigrations(pool);
}, 60_000);

afterAll(async () => {
  await pool.end();
  await container.stop();
});

it("round-trips through PostgreSQL", async () => {
  const result = await pool.query<{ value: string }>("SELECT 'ready' AS value");
  expect(result.rows[0]?.value).toBe("ready");
});
```

### Fastify Transport Tests

Build the application through a test composition function and use Fastify `app.inject()` without a real listening socket. Assert validation rejection, authn/authz, content type, status, headers, exact problem details, success body, and log/trace correlation.

Keep core behavior behind a fake port in transport tests; prove the real database adapter in its integration suite. Add a small composition smoke test with real registration order to catch plugin and hook mistakes.

### Contracts And Serialization

Boundary tests start from `unknown` and exercise Zod parsing. Cover missing, null, unexpected, oversized, malformed, boundary, and forward-compatible values. Do not construct already-typed objects and claim the parser works.

Golden fixtures pin important HTTP and event shapes. Keep them small, synthetic, reviewed, and regenerated only through an explicit command. A fixture update is a contract change.

For libraries, `npm pack` and install the tarball into a consumer fixture. Test documented imports, declarations, ESM runtime behavior, and supported Node versions.

### Determinism

A test dependent on scheduling, wall time, random order, network access, locale, timezone, or shared process state is not proof.

- Inject a fixed clock for decisions and use Jest fake timers only for timer orchestration.
- Never use real sleeps to “let work happen.” Await the owned promise or controlled event.
- Seed random generators and retain failing seeds.
- Sort unordered values before comparison when order is not contractual.
- Set locale and timezone explicitly for relevant suites.
- Avoid mutation of `process.env`; when unavoidable, restore it and do not run those tests concurrently.
- Use unique database identifiers and deterministic cleanup.

Jest workers run tests concurrently across files. Tests must not share ports, mutable singletons, database rows, fake timers, or global handlers. Reduce workers only for a measured infrastructure constraint, not to hide isolation bugs.

### Async And Cancellation Proof

Test pre-aborted signals, abort during work, timeout, retry exhaustion, sibling failure, bounded fan-out, queue overflow, and graceful drain. Use controlled promises to hold work at a known point and assert the observed concurrency maximum.

Every long-lived server, consumer, timer, worker, subscription, and pool has teardown proof. Run with open-handle detection during diagnosis; do not normalize leaked handles by forcing process exit.

### Negative Paths

Every changed trust boundary includes rejection tests. Security-sensitive code covers missing and invalid authentication, forbidden authorization, injection payloads, SSRF destinations, oversized input, replay, and redaction as applicable.

Data-sensitive code covers partial write, duplicate delivery, conflict, cancellation, rollback, retention, and deletion. Error mapping covers known typed errors and safe unknown fallback.

### Coverage Stance

Coverage is a map and a ratchet, not the goal. Require coverage of domain decisions, parsing, authorization, error/status mapping, and failure paths. Do not write contrived tests for generated code, composition-only wiring, or trivial accessors to inflate a percentage.

Collect coverage separately from `npm run verify` unless the repository has an established stable threshold. Prevent meaningful regression through diff review or a recorded ratchet. Branch coverage matters more than a high statement number when finite outcomes exist.

Do not exclude difficult code merely to improve the report. Refactor it behind a deterministic seam or document why a boundary is proven by a different suite.

### Property, Fuzz, Mutation, And Performance Tests

Property-based tests are useful for round trips, normalization idempotence, ordering laws, and state machines with large input spaces. Adding a generator library follows dependency review. Preserve discovered counterexamples as focused regression tests.

Fuzz custom parsers and decoders that accept untrusted bytes or strings. Bound input size and run exploration on demand or on a schedule, not in the fast gate.

Mutation testing is optional for critical domain or security modules when it reveals weak assertions. Benchmarks and load tests support explicit performance or capacity claims; they do not run as noisy merge gates without a controlled environment.

### Snapshots

Use snapshots only for stable, reviewable structured output where a diff communicates the contract better than focused assertions. Inline snapshots are preferred for small values. Large React DOM snapshots and automatic snapshot updates are forbidden.

Review every snapshot change. Do not snapshot secret-bearing configuration, real production payloads, timestamps, random IDs, or unstable framework markup.

### CI Placement

`npm run verify` runs deterministic unit and component Jest suites offline. The explicit integration job provisions pinned PostgreSQL, runs Testcontainers-backed suites, and publishes useful failure artifacts.

Critical end-to-end, load, fuzz, and mutation suites run at a frequency matching their cost and risk. A scheduled suite failure is owned work, not informational noise.

## Common Mistakes And Forbidden Patterns

- Pretending Jest 30 ESM is transparent or omitting its experimental VM-modules requirement.
- Tests that pass under a transformer while the production ESM artifact cannot start.
- Mocking `pg` and calling it database integration.
- React tests querying implementation classes or asserting hook internals.
- MSW handlers that accept any request or unhandled requests that fall through.
- Real sleeps, live network calls, shared mutable fixtures, or dependence on test order.
- Happy-path-only tests at parsing, auth, persistence, messaging, or cancellation boundaries.
- Snapshot walls, automatic snapshot approval, or aggregate coverage treated as correctness.
- Containers, servers, pools, timers, or workers left open after a suite.
- Retrying flaky tests until CI turns green.

## Verification And Proof

- Jest configuration and transformer versions are pinned and documented; the selected ESM path works from a clean checkout.
- Backend `dist` or the Vite production bundle passes a runtime smoke test outside Jest transforms.
- Unit and component suites pass offline, in random file order where supported, with no leaked handles.
- RTL tests use accessible queries and awaited user-event interactions.
- MSW rejects unhandled requests and covers malformed, error, abort, and retry behavior.
- Testcontainers suites apply real migrations and prove queries against pinned PostgreSQL.
- Negative tests cover every changed trust, authorization, data, and failure boundary.
- Async tests prove timeout, abort, concurrency bounds, and drain without real sleeps.
- Coverage review shows critical decisions and error branches are exercised without vanity exclusions.
- CI runs `npm run verify` plus the explicitly gated integration suite.

Related: [../foundations/async-and-cancellation.md](../foundations/async-and-cancellation.md), [../services/database.md](../services/database.md), and [../operations/ci-and-release.md](../operations/ci-and-release.md). Tool anchors: [Jest 30 ESM](https://jestjs.io/docs/30.0/ecmascript-modules), [Testing Library query priority](https://testing-library.com/docs/queries/about/#priority), [MSW](https://mswjs.io/docs/), and [Testcontainers for Node.js](https://node.testcontainers.org/).
