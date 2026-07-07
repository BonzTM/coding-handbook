# Web Applications (Server-Rendered)

Defaults for services that return HTML to browsers: templates, static assets, sessions, CSRF, and browser-facing security — layered on top of the HTTP service shape.

## Default Approach

A server-rendered web app **is** an HTTP service plus an HTML boundary. Everything in [http-services.md](http-services.md) still applies: the handler contract, server hardening timeouts, middleware order, pagination, and the error-envelope rules for any JSON endpoints it also exposes. This doc covers only what the HTML boundary adds.

### Scope

This shape is stdlib-first server rendering: `html/template` pages, embedded static assets, cookie sessions. It deliberately does **not** cover single-page apps or JavaScript build pipelines — a SPA is a separate frontend artifact talking to a JSON API per [http-services.md](http-services.md), and its toolchain lives outside the Go repo. If the spec says "web app" without qualification, server-rendered is the default; reach for a SPA only when the spec demands rich client-side interactivity, and record that as an ADR.

### Minimal Layout

```text
internal/api/web/
  server.go        # mux construction and middleware wiring
  handlers.go      # page handlers on a server struct
  templates.go     # embed.FS + parse-once template set
  templates/       # *.tmpl page and partial templates
  static/          # css, images, small js (embedded)
```

A repo serving both HTML pages and a JSON API keeps them as sibling packages (`internal/api/web`, `internal/api/http`) sharing one `internal/core`.

### Templates

- Use `html/template`, never `text/template`, for anything emitted to a browser — contextual auto-escaping is the primary XSS defense and only `html/template` has it.
- Embed templates with `embed.FS` and parse once at startup via `template.Must(template.ParseFS(...))` — a broken template fails the boot, not the request. In dev, a config-gated reparse-per-request is acceptable; the production path stays parse-once.
- Pass a typed per-page data struct to each template, not `map[string]any`. The struct is the page's contract; name it next to the handler.
- Build pages from a base layout plus partials (header, nav, flash). Keep logic in Go; templates render, they do not decide.
- Reach for `template.HTML` only on values the server itself constructed from trusted parts. Wrapping user input in `template.HTML` is the forbidden pattern that reintroduces XSS.

### Static Assets

- Embed assets and serve them with `http.FileServerFS` over the embedded tree; the binary stays the single deploy artifact.
- Cache HTML as `no-cache` (revalidate every navigation) and asset paths that change on deploy as long-lived: fingerprint asset URLs (a content hash or the build version in the path) and serve them with `Cache-Control: public, max-age=31536000, immutable`. Never serve mutable content under an immutable URL.
- Set explicit `Content-Type` where the file server cannot infer it, and keep `X-Content-Type-Options: nosniff` on (see Security Headers).

### Forms And The PRG Pattern

- Parse with `r.ParseForm()` under the body-size cap from [http-services.md](http-services.md); validate at the boundary and re-render the form with field errors and the user's input preserved — a validation failure is a `200` re-render or `422`, not a redirect that loses their work.
- Successful state-changing posts follow Post/Redirect/Get: `303 See Other` to the result page, so refresh never replays the write.
- One-shot notices ("saved", "deleted") ride a flash message in the session, set before the redirect and cleared on first read.

### Sessions

- Default: server-side sessions via `github.com/alexedwards/scs/v2` — session data lives in a store (its Postgres store in repos that already run Postgres; the in-memory store for tests/dev), and the cookie carries only the session token, revocable server-side.
- Cookie flags are explicit, not defaulted: `Secure: true`, `HttpOnly: true`, `SameSite=Lax` (`Strict` where the UX tolerates it). Set them in one place in `server.go`.
- Regenerate the session token on login and destroy the session on logout — fixation and stale-privilege defense.
- Sessions hold identity and small UI state (flash messages), not domain data. If it belongs in the database, it does not belong in the cookie store.
- Do not use JWTs in cookies as a session substitute: no revocation, growing claims, and a signature check is not a logout.

### CSRF

- Default: the stdlib `net/http.CrossOriginProtection` (Go 1.25+) wrapped around the mux — it rejects cross-origin state-changing requests using the browser's `Sec-Fetch-Site` header, falling back to `Origin` comparison. Add legitimate cross-origin callers explicitly via `AddTrustedOrigin`; exempt webhook-style endpoints explicitly, not by disabling the middleware.
- Escalate to token-based CSRF middleware (hidden form field) only when the app must support clients that send neither `Sec-Fetch-Site` nor `Origin` — that is an ADR per [../decisions/framework-selection.md](../decisions/framework-selection.md).
- CSRF protection sits with the edge middleware, before session and handler logic, and applies to every non-safe method. GET handlers must not mutate state — that invariant is what makes the whole model sound.

### Security Headers

Set once, in middleware, for every HTML response:

- `Content-Security-Policy` — start at `default-src 'self'; frame-ancestors 'none'` and loosen per directive with justification; a real CSP is the second XSS layer behind template escaping.
- `X-Content-Type-Options: nosniff`
- `Referrer-Policy: strict-origin-when-cross-origin`

A same-origin web app needs **no CORS middleware** — the CORS stance in [http-services.md](http-services.md) applies unchanged.

### HTML Error Pages

Browser-facing errors render a friendly error page (styled 404, generic 500) instead of the JSON envelope; JSON endpoints in the same binary keep the envelope. The logging rule is unchanged from [../foundations/errors-and-logging.md](../foundations/errors-and-logging.md): log once at the boundary, never leak internal detail (stack traces, SQL, dependency names) into the page body.

## Common Mistakes And Forbidden Patterns

- `text/template` (or hand-concatenated strings) rendering browser output — no contextual escaping, guaranteed XSS surface.
- `template.HTML` applied to user-influenced values to "fix" escaping.
- Parsing templates per request in production, or ignoring the parse error until the first page load.
- A state-changing GET handler — it breaks CSRF protection, caching, and prefetch safety at once.
- Session cookies without `Secure`/`HttpOnly`, or a session token that survives login/logout transitions.
- Domain data accumulating in the session store instead of the database.
- Fingerprint-cached URLs whose content can change, or HTML served with long-lived cache headers.
- Disabling CSRF protection app-wide because one webhook endpoint needed an exemption.
- A JSON error envelope rendered to a browser page, or a stack trace rendered into HTML.

## Verification And Proof

- handler tests with `httptest` asserting rendered HTML fragments (golden files for stable pages) and the `Content-Type`
- a template-parse test: the full `ParseFS` set loads without error (this also runs implicitly at startup)
- negative tests: a cross-origin POST without `Sec-Fetch-Site`/`Origin` trust is rejected; a state-changing GET does not exist
- cookie-flag assertions on the session cookie (`Secure`, `HttpOnly`, `SameSite`)
- an XSS probe test: user input containing `<script>` round-trips escaped in the rendered page
- `make verify` green, per the repo baseline
