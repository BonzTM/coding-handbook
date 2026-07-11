# Solution And Project Design

Project boundaries, accessibility rules, and dependency direction for .NET code that stays maintainable under growth.

## Default Approach

Prefer few projects with compiler-enforced boundaries and a minimal public surface. Project references are the boundary mechanism: a folder convention only suggests a rule, a missing `<ProjectReference>` makes the violation a build error. The default service is exactly three projects — `Orders.Api`, `Orders.Core`, `Orders.Infrastructure` — plus tests (see [project-setup.md](project-setup.md)).

### Dependency Direction

| Project | Can reference | Must not reference |
|---|---|---|
| `Orders.Api` | `Orders.Core` everywhere; `Orders.Infrastructure` only from the composition root (`Program.cs` and the `Add*` wiring it calls) | `Orders.Infrastructure` types from endpoints, filters, or request handling |
| `Orders.Core` | the BCL and abstraction-only packages (e.g. `Microsoft.Extensions.Logging.Abstractions`; `TimeProvider` is stdlib) | ASP.NET Core, EF Core, broker SDKs, any other project in the solution |
| `Orders.Infrastructure` | `Orders.Core` (to implement its ports), EF Core and the Npgsql provider, external client SDKs | `Orders.Api`, anything transport-shaped |
| `Orders.UnitTests` / `Orders.IntegrationTests` | the projects under test; internals via `InternalsVisibleTo` | nothing references test projects, ever |

`Orders.Api` holding a reference to `Orders.Infrastructure` exists solely so `Program.cs` can bind implementations to Core ports. An endpoint that touches a `DbContext` or a typed client directly has broken the boundary even though it compiles.

### Interface Placement

- Define interfaces where they are consumed, not where they are implemented. In the three-project shape the consumer of persistence and outbound behavior is `Orders.Core`, so the ports live there and `Orders.Infrastructure` implements them.
- Return concrete types by default; accept interfaces when a caller needs substitution.
- Keep interfaces small: 1–3 members. A wide interface is a class wearing a disguise; either split it or pass the concrete type.
- Name the interface for what the consumer needs (`IOrderRepository`, `IPaymentGateway`), not for what the implementer is (`IEfOrderStore`, `IStripeClientWrapper`).
- Do not create an interface for every class by reflex. The trigger is a real second implementation or a test seam that a hand-rolled fake needs (see [../quality/testing.md](../quality/testing.md)); until then, use the concrete type.

```csharp
namespace Orders.Core;

public interface IOrderRepository
{
    Task<Order?> GetAsync(OrderId id, CancellationToken cancellationToken);
    Task AddAsync(Order order, CancellationToken cancellationToken);
}
```

### Accessibility And InternalsVisibleTo

- Start `internal` (the C# default for top-level types). Make a type `public` only when another project in the solution truly consumes it.
- In an application solution, `public` means "the other two projects may use this"; in a library, `public` is a support commitment with compatibility cost (see [contracts-and-compatibility.md](contracts-and-compatibility.md)). Either way, the smallest surface that serves real callers wins.
- `InternalsVisibleTo` is allowed for exactly one purpose: a project's matching test project. Declare it in the `.csproj`:

```xml
<ItemGroup>
  <InternalsVisibleTo Include="Orders.UnitTests" />
</ItemGroup>
```

- `InternalsVisibleTo` between production projects is forbidden. If Api needs an internal from Infrastructure, the boundary is wrong — move the type or make the seam an explicit Core port.

### Namespaces And Naming

- Namespaces mirror folders exactly, and every file uses a file-scoped namespace. Root namespace equals project name: code in `src/Orders.Core/Pricing/` lives in `namespace Orders.Core.Pricing;`.
- Because `using` directives strip namespaces at the call site, type names must stand alone: `PricingCalculator` reads correctly in any file; a bare `Calculator` does not. This is the inverse of qualified-by-package languages — do not shorten type names on the assumption the namespace will disambiguate.
- Namespaces are named for purpose, not mechanics: `Orders.Core.Pricing`, not `Orders.Core.Utils`, `Helpers`, `Common`, `Models`, or `Services`.

### Generics And Type Parameters

Reach for type parameters only for type-safe containers and algorithms that would otherwise duplicate code per type or fall back to `object` plus casts. Use ordinary interfaces for behavioral polymorphism — when callers vary by what a value *does*, not by what type it *holds*.

- Decision rule: generics when the type is the only thing that changes; interfaces when behavior changes.
- Do not add type parameters speculatively, and never to dress up a single concrete type. Write the concrete version; generalize on the second real caller.
- Constrain to the minimum: `notnull` for dictionary keys, `IEquatable<T>`/`IComparable<T>` for comparison, `INumber<T>` only when generic arithmetic is genuinely the point. An unconstrained `T` that the body immediately casts is a bug in the design.
- The classic failure is `IRepository<T>` with generic CRUD: it flattens every aggregate to the same shape and hides real query needs. Each aggregate gets its own narrow port (see [../services/database.md](../services/database.md)).
- Overly generic APIs hurt readability more than the duplication they remove. If the signature needs a paragraph to explain, prefer two concrete methods. See [data-modeling.md](data-modeling.md) for choosing the underlying types these operate on.

### Constructor Design

Constructor injection with explicit dependencies is the composition mechanism — no service locator, no `IServiceProvider` passed around. When a component has optional or growing configuration, bind an options class rather than adding boolean or primitive parameters:

```csharp
public sealed class DispatchOptions
{
    public int MaxAttempts { get; init; } = 3;
    public TimeSpan BaseDelay { get; init; } = TimeSpan.FromMilliseconds(200);
}

public sealed class OrderDispatcher(
    IOrderPublisher publisher,
    IOptions<DispatchOptions> options,
    TimeProvider timeProvider,
    ILogger<OrderDispatcher> logger)
{
    private readonly DispatchOptions _options = options.Value;
    // ...
}
```

- Required collaborators are constructor parameters; tuning knobs live in the options class, validated at startup (see [configuration.md](configuration.md)).
- A long constructor is acceptable when it reveals real dependencies; hiding them behind a container facade or ambient statics is not cleaner, it is invisible.
- For types not built by the DI container, pass a plain options record explicitly — the shape is the same, minus `IOptions<T>`.

### File Organization

- One top-level type per file, file named after the type: `Order.cs`, `OrderDispatcher.cs`, tested by `OrderTests.cs` in the matching test project.
- Split files by responsibility, not by type kind — no `Interfaces.cs`, `Models.cs`, or `Dtos.cs` junk drawers. The one sanctioned grouping: the small request/response records for a single endpoint group may live in that group's contracts file, because they change together.
- The canonical `Orders.Api` layout is `Endpoints/` (one file per endpoint group) plus `Telemetry/`; cross-cutting concerns that outgrow an endpoint filter get their own file, same rule (see [shared-constructs.md](shared-constructs.md) and [../services/http-services.md](../services/http-services.md)).

## Common Mistakes And Forbidden Patterns

- Layer-cake solutions: `Orders.Domain` + `Orders.Application` + `Orders.Abstractions` + `Orders.Contracts` + `Orders.Persistence` for one service. Three projects is the default shape, not the floor.
- A `Shared` or `Common` project created to break a circular reference, which then becomes the real architecture.
- `IRepository<T>` / generic unit-of-work abstractions layered over EF Core, which already is a unit of work.
- An interface for every class before a second implementation or test seam exists.
- Transport DTOs or EF entities leaking into `Orders.Core`.
- `InternalsVisibleTo` between production projects.
- Service-locator patterns: injecting `IServiceProvider`, static `ServiceLocator.Get<T>()`, or resolving from `HttpContext.RequestServices` in business code.
- Speculative generics: type parameters added before a second concrete type exists.
- Everything `public` by reflex, so the compiler can no longer tell you what the real surface is.

## Verification And Proof

- `dotnet build` (via `pwsh ./verify.ps1`) proves the reference graph: Core cannot reach what it does not reference, and cycles do not compile.
- Grep for `InternalsVisibleTo` — every hit names a matching test project and nothing else.
- Review `Orders.Core`'s public surface: every public type is a commitment to the other two projects; it should look intentional rather than accidental.
- For each type parameter, name the concrete duplication or cast-heavy code it removes; if you cannot, drop it.
- For each interface, count the members (target 1–3) and confirm it sits in the consuming project.
- Review a proposed new project by asking: what owns this behavior, who references it, and what contract does that create?
