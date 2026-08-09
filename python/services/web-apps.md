# Web Applications (Server-Rendered)

Browser-facing HTML, templates, static assets, sessions, CSRF, and security policy layered on the FastAPI service shape.

## Default Approach

A server-rendered web app is an [HTTP service](http-services.md) plus an HTML/browser boundary. Use FastAPI, Jinja2, explicit view models, and the same framework-free core. A SPA or separate frontend toolchain requires a distinct architecture decision.

### Layout

```text
src/<app>/api/http/
  pages/                 # page routers and form adapters
  templates/             # Jinja2 layouts, pages, partials
  static/                # CSS, images, small reviewed JavaScript
  sessions.py            # server-side session adapter
  csrf.py                # synchronizer-token generation/verification
```

JSON and page routers remain separate adapters even when one FastAPI app serves both. They share core use cases, not transport DTOs.

### Jinja2 Templates

Create one `Jinja2Templates` environment during app construction. Enable autoescape for HTML/XML templates with `select_autoescape`, use `StrictUndefined` so missing view data fails visibly, and parse representative templates at startup/test time. Jinja documents autoescape selection in its [API guide](https://jinja.palletsprojects.com/en/stable/api/#jinja2.select_autoescape).

Pass a dedicated typed view model converted to the template context at the adapter. Templates render; core decides. Keep business rules, database access, authorization decisions, and network calls out of filters/globals/templates.

Autoescape stays on. `|safe`, `Markup`, or HTML-bearing values require security review proving the content is server-constructed or sanitized by a vetted HTML sanitizer for an explicitly allowed rich-text contract. Never mark user, database, query, or dependency text safe to “fix” display.

### Static Files

Mount FastAPI/Starlette `StaticFiles` at an explicit path and keep assets separate from templates. Fingerprint immutable asset names/content during the build and serve them with long-lived immutable caching; HTML and mutable assets revalidate. Never put uploaded/user-controlled files in the executable static tree.

Set correct content types and `X-Content-Type-Options: nosniff`. If assets move to a CDN, update CSP/CORS/cache policy through an ADR and prove integrity/deployment behavior.

### Forms And PRG

Bound form bodies before parsing and validate with an adapter-owned Pydantic/form model. On validation failure, render the form with safe entered values and field errors; do not redirect and lose context.

After a successful state-changing POST, use `303 See Other` to a GET representation. This Post/Redirect/Get flow prevents browser refresh from replaying the form. A one-time flash message lives in the server-side session and is consumed once.

GET, HEAD, and OPTIONS never change state. Authorization is checked again on the state-changing handler; a hidden field is not trusted state.

### Sessions

Use an opaque high-entropy session identifier in the cookie and store session state server-side in an existing durable store. The cookie is `Secure`, `HttpOnly`, `SameSite=Lax` by default, host-only where possible, has the narrowest path, and has explicit idle/absolute expiry. Rotate the identifier on authentication/privilege change and destroy it on logout.

Session state holds identity references, CSRF token material, and small UI state such as flashes; domain records stay in their source of truth. Never place secrets, PII, authorization claims that cannot be revoked, or whole domain objects in client-side session data.

Starlette's built-in `SessionMiddleware` stores signed cookie data that is readable by the client, as its [middleware documentation](https://www.starlette.io/middleware/#sessionmiddleware) states. It is not the default for authenticated server-side sessions. A signed cookie proves integrity, not confidentiality, revocation, or server-side expiry.

### CSRF

FastAPI has no built-in CSRF protection. Implement an adapter-owned synchronizer-token pattern tied to the server-side session, following the [OWASP CSRF guidance](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html#synchronizer-token-pattern):

1. generate an unpredictable token with `secrets` when the session is established/rotated
2. store token material server-side in that session
3. render the token in a hidden field for every state-changing form, never a URL
4. compare the submitted token in constant time
5. reject missing/mismatched tokens before business logic

Apply protection to every cookie-authenticated unsafe method, including JSON/AJAX routes; use a custom header for same-origin JavaScript. Validate `Origin` and Fetch Metadata (`Sec-Fetch-Site`) as defense in depth with an explicit proxy/legacy-browser policy. `SameSite` is defense in depth, not the complete control. Webhooks and non-cookie machine APIs use separate authentication and a narrow explicit exemption, never a global disable.

### Security Headers And CSP

Set headers once for every HTML response:

- `Content-Security-Policy`: start with `default-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'`; add nonce/hash-based script/style directives and external origins only with review.
- `X-Content-Type-Options: nosniff`.
- `Referrer-Policy: strict-origin-when-cross-origin` or stricter when the product allows.
- HSTS only at the HTTPS-owning edge and only after subdomain/preload consequences are reviewed.
- sensitive pages use `Cache-Control: no-store`; ordinary HTML uses explicit revalidation policy.

Do not use `unsafe-inline` as the standing CSP. Keep JavaScript in static assets; if inline bootstrapping is unavoidable, use a per-response nonce and test it. OWASP describes CSP as a second layer, not a substitute for output encoding, in its [CSP guidance](https://cheatsheetseries.owasp.org/cheatsheets/Content_Security_Policy_Cheat_Sheet.html).

### XSS Rules

Treat every value from users, URLs, databases, dependencies, and translations as untrusted. Let Jinja perform context-aware escaping in HTML/attribute contexts. Do not interpolate untrusted values directly into script, style, event-handler, URL-scheme, or raw HTML contexts; OWASP calls these dangerous contexts in its [XSS prevention guidance](https://cheatsheetseries.owasp.org/cheatsheets/Cross_Site_Scripting_Prevention_Cheat_Sheet.html#dangerous-contexts).

Sanitization is required only when the product deliberately accepts rich HTML and uses a vetted allowlist sanitizer. Validation alone does not make HTML safe. Client-side DOM construction uses safe text/property sinks, never `innerHTML` with untrusted data.

### Browser Error Pages

Render friendly, generic HTML for browser 404/403/500 paths. JSON routes retain `application/problem+json`. Unknown errors include a request identifier but no exception, SQL, dependency, secret, or stack trace. Log once at the acting boundary.

## Common Mistakes And Forbidden Patterns

- String-built HTML, autoescape disabled, or `|safe` applied to untrusted/unsanitized content.
- Template filters performing database/network I/O or making authorization decisions.
- User uploads served from the executable static tree or mutable content under immutable cache names.
- State-changing GET, successful POST rendered directly without PRG, or unbounded form body.
- Sensitive/auth claims stored in readable signed cookies; session ID not rotated on login/logout.
- SameSite, CORS, Origin, or Fetch Metadata treated as the only CSRF control.
- CSRF token in URL/log/cookie for the synchronizer pattern, non-constant comparison, or global webhook exemption.
- CSP disabled or permanently weakened with `unsafe-inline`; HSTS set without owning HTTPS.
- Stack trace/internal detail in an HTML error page.

## Verification And Proof

```bash
uv run pytest -k "template or page or session or csrf or xss"
make verify
```

Prove template rendering and missing-variable failure, `<script>`/attribute/URL-context XSS probes, `|safe` review inventory, static content types/cache policy, PRG status/location, session rotation/destruction/expiry/cookie flags, CSRF success plus missing/wrong/cross-session/cross-origin failures, webhook exemption scope, CSP/security headers, sensitive-page caching, and generic error pages. Run one browser-oriented smoke test through the actual HTTPS edge when cookies or HSTS depend on it.

Related: [HTTP services](http-services.md), [serialization](../foundations/serialization.md), [security](../operations/security.md), and [testing](../quality/testing.md).
