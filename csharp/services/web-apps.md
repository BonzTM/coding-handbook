# Web Applications (Server-Rendered)

Defaults for services that return HTML to browsers: Razor Pages, static assets, auth cookies, antiforgery, and browser-facing security — layered on top of the HTTP service shape.

## Default Approach

A server-rendered web app **is** an HTTP service plus an HTML boundary. Everything in [http-services.md](http-services.md) still applies: the endpoint contract, hardening timeouts, middleware order, pagination, and the ProblemDetails rules for any JSON endpoints it also exposes. This doc covers only what the HTML boundary adds.

### Scope

This shape is Razor Pages on the same host: server-rendered `.cshtml` pages, `wwwroot` static assets, cookie authentication. It deliberately does **not** cover single-page apps or JavaScript build pipelines — a SPA is a separate frontend artifact talking to a JSON API per [http-services.md](http-services.md), and its toolchain lives outside the .NET repo. If the spec says "web app" without qualification, Razor Pages is the default. Blazor (Server or WebAssembly) is not the default — it earns its place only when the spec demands rich stateful interactivity, and that is an ADR per [../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md) and [../decisions/framework-selection.md](../decisions/framework-selection.md). HTMX-style progressive enhancement over Razor Pages is allowed and is the first answer to "we need some interactivity".

### Minimal Layout

```text
src/Orders.Api/
  Pages/
    _ViewImports.cshtml          # @namespace, @addTagHelper
    _ViewStart.cshtml            # Layout = "_Layout"
    Shared/
      _Layout.cshtml             # base layout: head, nav, flash, @RenderBody()
      _Flash.cshtml              # partials: header, nav, flash
    Orders/
      Index.cshtml               # page + PageModel pairs per resource
      Index.cshtml.cs
      Create.cshtml
      Create.cshtml.cs
    Error.cshtml
  wwwroot/                       # css, images, small js — served via MapStaticAssets
```

A repo serving both HTML pages and a JSON API keeps them on one host: `Pages/` for HTML, `Endpoints/` for JSON, sharing one `Orders.Core`.

### Razor Pages And Layout

- Razor's `@` expressions HTML-encode by default — contextual escaping is the primary XSS defense; keep it. `Html.Raw` (and `HtmlString` around user input) is the forbidden pattern that reintroduces XSS: reach for it only on values the server itself constructed from trusted parts.
- The `PageModel` is the page's typed contract: expose typed properties (`IReadOnlyList<OrderSummary> Orders`), never `ViewData`-bag soup, and keep decisions in the model or core — views render, they do not decide.
- Build pages from one `_Layout.cshtml` plus partials (`_Nav`, `_Flash`); `_ViewStart.cshtml` applies the layout, `_ViewImports.cshtml` holds the tag-helper and namespace imports once.
- A broken page fails at build, not first render: keep `<RazorCompileOnBuild>` on its default (build-time compilation) so template errors surface in `pwsh ./verify.ps1`, not in production.

```csharp
public sealed class IndexModel(IOrderService orders) : PageModel
{
    public IReadOnlyList<OrderSummary> Orders { get; private set; } = [];

    public async Task OnGetAsync(CancellationToken cancellationToken) =>
        Orders = await orders.ListRecentAsync(cancellationToken);
}
```

### Static Assets

- Serve `wwwroot` with `app.MapStaticAssets()` (and `.WithStaticAssets()` on the Razor Pages mapping): assets are compressed at build time, content-fingerprinted, and served with `Cache-Control: immutable` plus ETags — the fingerprint pipeline that had to be hand-built in other stacks is the framework default here.
- Reference assets through the built-in `<link>`/`<script>`/`<img>` tag helpers with `~/` paths so the fingerprinted URLs are emitted for you. Never hardcode a fingerprinted filename, and never serve mutable content under an immutable URL.
- HTML responses stay `no-cache` (revalidate every navigation); only fingerprinted asset URLs are long-lived.
- Keep `X-Content-Type-Options: nosniff` on every response (see Security Headers).

### Forms And The PRG Pattern

- Bind with `[BindProperty]` to a form DTO carrying DataAnnotations; on `!ModelState.IsValid`, `return Page()` — re-render with field errors (`asp-validation-for`) and the user's input preserved. A validation failure is a re-render, not a redirect that loses their work.
- Successful state-changing posts follow Post/Redirect/Get: `RedirectToPage` to the result page, so refresh never replays the write.
- One-shot notices ("saved", "deleted") ride `TempData` — set before the redirect, cleared on first read, rendered by `_Flash.cshtml`.
- The body-size cap and validation stance from [http-services.md](http-services.md) apply to form posts unchanged.

```csharp
public sealed class CreateModel(IOrderService orders) : PageModel
{
    [BindProperty]
    public CreateOrderForm Form { get; set; } = new();

    public async Task<IActionResult> OnPostAsync(CancellationToken cancellationToken)
    {
        if (!ModelState.IsValid)
        {
            return Page();
        }

        var id = await orders.CreateAsync(Form.ToCommand(), cancellationToken);
        TempData["Flash"] = "Order created.";
        return RedirectToPage("Details", new { id });
    }
}
```

### Sessions And Authentication State

- Default: ASP.NET Core cookie authentication. The cookie carries an encrypted, signed auth ticket (identity + claims) protected by Data Protection — no server-side session store to run, no JWT-in-cookie hazards.
- When the app needs server-side revocation ("log out everywhere", instant privilege changes) or the claims set grows past cookie size, plug an `ITicketStore` (`CookieAuthenticationOptions.SessionStore`): the cookie shrinks to an opaque key and tickets live server-side, revocable — that is this stack's server-session escape hatch, not a reason to reach for `ISession`.
- Cookie flags are explicit, not defaulted, set in one place in `Program.cs`:

```csharp
builder.Services.AddAuthentication(CookieAuthenticationDefaults.AuthenticationScheme)
    .AddCookie(options =>
    {
        options.Cookie.Name = "__Host-orders";
        options.Cookie.SecurePolicy = CookieSecurePolicy.Always;
        options.Cookie.HttpOnly = true;
        options.Cookie.SameSite = SameSiteMode.Lax; // Strict where the UX tolerates it
        options.ExpireTimeSpan = TimeSpan.FromHours(8);
        options.SlidingExpiration = true;
        options.LoginPath = "/Account/Login";
    });
```

- `SignInAsync` on login issues a fresh ticket and `SignOutAsync` on logout clears it — fixation and stale-privilege defense. Call `SignInAsync` again after any privilege change so the ticket's claims match reality.
- The auth cookie holds identity and small UI state (flash via `TempData`), not domain data. If it belongs in the database, it does not belong in a cookie. Do not enable the `ISession`/`AddSession` middleware to smuggle domain state — that is the forbidden pattern here.
- Cookie auth, antiforgery, and `TempData` all ride the Data Protection key ring. Multi-instance deployments must persist and share the key ring, or every deploy logs out every user and breaks in-flight forms — see [../operations/deployment.md](../operations/deployment.md).

### Antiforgery

- Razor Pages ships CSRF defense on by default: the form tag helper (`<form method="post">`) injects the antiforgery token automatically, and the runtime validates it on every unsafe method (POST/PUT/DELETE) for pages. Do not turn this off; there is nothing to wire.
- Exempt webhook-style endpoints explicitly and individually (`[IgnoreAntiforgeryToken]` on that one page, or keep webhooks on JSON endpoints where token auth applies) — never by disabling validation app-wide.
- Minimal-API endpoints in the same host that accept form posts need `builder.Services.AddAntiforgery()` + `app.UseAntiforgery()` and an explicit token in the form; prefer keeping browser forms on Razor Pages where validation is automatic.
- GET handlers (`OnGet*`) must not mutate state — that invariant is what makes the whole antiforgery model sound.

### Security Headers

Set once, in middleware, for every HTML response:

- `Content-Security-Policy` — start at `default-src 'self'; frame-ancestors 'none'` and loosen per directive with justification; a real CSP is the second XSS layer behind Razor encoding.
- `X-Content-Type-Options: nosniff`
- `Referrer-Policy: strict-origin-when-cross-origin`

```csharp
app.Use(async (context, next) =>
{
    context.Response.Headers["Content-Security-Policy"] =
        "default-src 'self'; frame-ancestors 'none'";
    context.Response.Headers["X-Content-Type-Options"] = "nosniff";
    context.Response.Headers["Referrer-Policy"] = "strict-origin-when-cross-origin";
    await next(context);
});
```

A same-origin web app needs **no CORS middleware** — the CORS stance in [http-services.md](http-services.md) applies unchanged. `UseHsts` follows the transport-security rule there: only when this host terminates TLS.

### Progressive Enhancement

Interactivity is added HTMX-style: small JS in `wwwroot` posting to page handlers or fetching HTML fragments, with every page still functional without it. Named page handlers (`OnGetRowAsync` returning `Partial("_OrderRow", model)`) serve the fragments. The moment someone proposes a client-side router, a bundler, or component state on the server, that is a SPA or Blazor conversation — out of this doc's scope and into an ADR.

### HTML Error Pages

Browser-facing errors render a friendly error page instead of the JSON envelope: `app.UseExceptionHandler("/Error")` plus `app.UseStatusCodePagesWithReExecute("/Error/{0}")` for a styled 404 and a generic 500. JSON endpoints in the same host keep ProblemDetails. The logging rule is unchanged from [../foundations/errors-and-logging.md](../foundations/errors-and-logging.md): log once at the boundary, and never leak internal detail (stack traces, SQL, dependency names) into the page body — `UseDeveloperExceptionPage` behavior must never reach production.

## Common Mistakes And Forbidden Patterns

- `Html.Raw` (or `HtmlString`) applied to user-influenced values to "fix" encoding — guaranteed XSS surface.
- String-concatenated HTML in C# instead of Razor's contextual encoding.
- A state-changing `OnGet` handler — it breaks antiforgery, caching, and prefetch safety at once.
- Auth cookies without `Secure`/`HttpOnly`/`SameSite`, or a ticket that survives logout or privilege changes.
- Domain data accumulating in `TempData`, `ISession`, or ticket claims instead of the database.
- Disabling antiforgery validation app-wide because one webhook endpoint needed an exemption.
- An ephemeral Data Protection key ring on a multi-instance deployment — every deploy becomes a mass logout plus antiforgery failures.
- Hand-rolled asset fingerprinting or hardcoded fingerprinted filenames next to `MapStaticAssets`, or HTML served with long-lived cache headers.
- A JSON ProblemDetails body rendered to a browser page, or a stack trace rendered into HTML.
- Reaching for Blazor because it is "the .NET way" without the ADR the escape hatch requires.

## Verification And Proof

- page tests through `WebApplicationFactory<Program>` asserting rendered HTML fragments and `Content-Type: text/html`
- an antiforgery negative test: a POST without the token is rejected (`400`), and the form tag helper emits the hidden token field
- cookie-flag assertions on the auth cookie (`Secure`, `HttpOnly`, `SameSite`, `__Host-` prefix)
- an XSS probe test: user input containing `<script>` round-trips encoded in the rendered page
- a PRG test: a valid POST answers with a redirect, an invalid POST re-renders with field errors and preserved input
- negative test: no state-changing GET handler exists (route audit)
- run `pwsh ./verify.ps1` — restore (locked), format-check, build (warnings-as-errors), test, audit; Razor build-time compilation makes template errors fail the build stage
