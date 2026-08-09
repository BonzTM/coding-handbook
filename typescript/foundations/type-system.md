# Type System

Strict TypeScript rules for making invalid states difficult to construct and boundary failures explicit.

## Default Approach

Use the compiler to prove internal relationships after runtime schemas have established trust.

### Strict Baseline

Enable the full strict baseline from [project-setup.md](project-setup.md). Type-aware ESLint is part of the contract. Do not weaken a repository-wide option to accommodate one dependency or file; isolate and document the exception.

`unknown` is the input type for untrusted values. Parse or narrow it before use. `any` is forbidden except at a proven third-party interop boundary with a local wrapper and suppression rationale.

### Narrowing And Exhaustiveness

- Prefer discriminated unions for finite state and result models.
- Use exhaustive `switch` handling enforced by lint.
- Write predicate or assertion functions only when they perform a real runtime check.
- Prefer `satisfies` when checking a literal while preserving its useful inferred type.
- Use `never` to prove impossible branches, not to conceal missing behavior.
- Do not use truthiness when `0`, `false`, or an empty string is meaningful.

Model finite state with a discriminant and make the impossible branch compile-visible:

```ts
type PaymentState =
  | Readonly<{ kind: "pending"; submittedAt: Date }>
  | Readonly<{ kind: "settled"; receiptId: string }>
  | Readonly<{ kind: "failed"; reason: string }>;

function assertNever(value: never): never {
  throw new Error("unhandled payment state", { cause: value });
}

export function describePayment(state: PaymentState): string {
  switch (state.kind) {
    case "pending":
      return `submitted ${state.submittedAt.toISOString()}`;
    case "settled":
      return `receipt ${state.receiptId}`;
    case "failed":
      return state.reason;
    default:
      return assertNever(state);
  }
}
```

Use `satisfies` to check a complete literal without widening away useful values:

```ts
type RetryClass = "never" | "transient";
type Dependency = "billing" | "inventory";

const retryPolicy = {
  billing: "never",
  inventory: "transient",
} as const satisfies Readonly<Record<Dependency, RetryClass>>;
```

### Assertions And Non-Null Values

Type assertions do not validate data. Replace `as T` at trust boundaries with Zod parsing. A narrow assertion is acceptable only where a runtime invariant is established locally and TypeScript cannot express it; comment the invariant when it is not obvious.

Non-null assertions are forbidden in production code. Narrow through control flow, validate construction, or model absence explicitly. Indexed access remains possibly undefined under `noUncheckedIndexedAccess`.

### Generics And Structural Types

Add a generic only when callers preserve a real relationship between inputs and outputs. Avoid generic parameter lists that merely rename `unknown`. Constrain the parameter as narrowly as the operation requires.

Structural compatibility is not semantic compatibility. Use branded identifiers or constructors where mixing same-shaped values would be dangerous. Keep brands module-owned and validate their underlying representation at boundaries.

The module that owns an identifier also owns its constructor:

```ts
import { z } from "zod";

declare const widgetIdBrand: unique symbol;
export type WidgetId = string & { readonly [widgetIdBrand]: true };

const widgetIdSchema = z
  .string()
  .uuid()
  .transform((value): WidgetId => value as WidgetId);

export function parseWidgetId(input: unknown): WidgetId {
  return widgetIdSchema.parse(input);
}
```

The assertion is contained after the UUID runtime check. Callers cannot construct the brand through a public unchecked cast.

### Unknown At Trust Boundaries

Decode external input to `unknown`, then parse it before field access:

```ts
import { z } from "zod";

const commandSchema = z.strictObject({
  name: z.string().trim().min(1).max(100),
  enabled: z.boolean(),
});

type Command = z.infer<typeof commandSchema>;

export function parseCommandJson(json: string): Command {
  const decoded: unknown = JSON.parse(json);
  return commandSchema.parse(decoded);
}
```

Do not annotate `JSON.parse` with the desired result type. The runtime schema establishes that type.

### Negative Type Proof

Use `@ts-expect-error` only in a compile-time test that proves misuse stays rejected:

```ts
declare function loadWidget(id: WidgetId): Promise<void>;

export async function widgetIdTypeProof(): Promise<void> {
  // @ts-expect-error plain strings are not validated WidgetId values
  await loadWidget("550e8400-e29b-41d4-a716-446655440000");

  const id = parseWidgetId("550e8400-e29b-41d4-a716-446655440000");
  await loadWidget(id);
}
```

The directive fails when the expected diagnostic disappears. It is proof, not a production suppression.

### Declaration And Augmentation

Prefer local wrappers to global declaration merging. Framework augmentation belongs in one named file, is included explicitly by tsconfig, and has a runtime initialization test proving the augmented property exists.

Published declarations are API artifacts. Review inferred public return types, avoid leaking dependency-private types, and run an install smoke test against packed output.

## Common Mistakes And Forbidden Patterns

- `any`, `@ts-ignore`, non-null assertions, double assertions, or `as unknown as T` used to silence design errors.
- Optional properties used where a discriminated state is required.
- Catch variables assumed to be `Error` without narrowing.
- Enums used where string-literal unions and runtime schemas provide the contract more clearly.
- Clever conditional types that make ordinary call sites or diagnostics unreadable.
- Global augmentation scattered across features.
- Compiler strictness reduced because dependency declarations are noisy; keep `skipLibCheck: true` instead.

## Verification And Proof

- `tsc --noEmit` passes with every strict baseline option enabled.
- ESLint reports no unsafe assignment, unsafe call, floating promise, or missing exhaustive case.
- Negative type tests prove rejected public API uses where compatibility is important.
- Boundary tests show malformed `unknown` values are rejected before domain use.
- Public declaration output contains no accidental private dependency types.
- Every suppression names the rule, the invariant, and the smallest affected line.

Related: [data-modeling.md](data-modeling.md), [serialization.md](serialization.md), and [../quality/linting.md](../quality/linting.md).
