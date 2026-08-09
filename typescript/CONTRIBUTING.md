# Contributing To This Handbook

> **Handbook-maintenance document.** This governs changes to the handbook itself, not to application repos. It is not part of the app-generation contract.

How to change the TypeScript handbook itself without breaking the contract it exists to enforce.

## Default Approach

This handbook is a control plane, not prose. A change is correct only when the fast path, routing, slow path, recipes, templates, and reference packages still agree after it lands.

### Know The Two-Speed Model You Are Editing

- **Fast path — [AGENTS.md](AGENTS.md).** Repo-wide invariants, change defaults, Change Routing, and verification. Refine it; never weaken it in passing.
- **Slow path — [maintainer-reference.md](maintainer-reference.md).** Architecture, module map, lifecycle, test taxonomy, and rationale.

Topic depth lives in `foundations/`, `services/`, `quality/`, and `operations/`; scaffolding in `templates/`; procedures in `recipes/`; gates in `checklists/`; binding decisions in `decisions/`.

### Use The House Templates

Match the existing shape exactly; do not invent a new one.

- **Topic docs:** `# Title` -> purpose -> `## Default Approach` -> `## Common Mistakes And Forbidden Patterns` -> `## Verification And Proof` -> optional related links.
- **Recipes:** Files To Touch / Steps / Invariants To Preserve / Proof, with concrete copy-pasteable commands.
- **Checklists:** traceable `- [ ]` boxes tied to evidence a reviewer can point at.
- **Templates:** fill-in skeletons with explicit `<PLACEHOLDER>` tokens, not finished project prose.
- **Indexes and glossary:** a clear navigational shape.

Voice is terse, opinionated, contract-not-tutorial. Cross-link only files that exist or land in the same change.

### Keep The Sync Surfaces In Sync

- **AGENTS.md <-> maintainer-reference.md <-> recipes.** A new invariant needs rationale and procedure; a recipe that introduces a rule needs routing.
- **Change Routing rows** point to files that exist and list the real also-update set.
- **Templates and reference packages** embody the prose. Change them with the rule they scaffold.
- **README reading paths, recipes index, and checklists index** enumerate the actual files.
- **Compiler, ESLint, Jest, package scripts, CI, Docker, and `.nvmrc`** must tell one runtime and verification story.

### Keep The Reference Packages Green

The three reference packages are executable proof. Run the gate from each changed or affected package:

```bash
(cd reference/exampleservice && npm run verify)
(cd reference/exampleworker && npm run verify)
(cd reference/examplefrontend && npm run verify)
```

Also validate template syntax, internal links, indexes, and byte-level config agreement directly. Never claim an exemplar gate was executed when it was not.

### Changing An Invariant Requires An ADR

Repo-wide invariants in [AGENTS.md](AGENTS.md) and Non-Negotiables in [README.md](README.md) change only through [architecture-decision-records.md](decisions/architecture-decision-records.md), using [templates/adr-template.md](templates/adr-template.md). After acceptance, move fast path, slow path, topic docs, recipes, templates, and reference packages together.

## Common Mistakes And Forbidden Patterns

- Editing one contract surface and leaving routing, rationale, recipe, template, or reference proof behind.
- Weakening Node, ESM, strict types, boundary parsing, architecture, or verification without an ADR.
- Inventing a document shape or tutorial voice.
- Cross-linking absent files or leaving a new file out of its directory index.
- Publishing unverified package versions, flags, config keys, commands, or URLs.
- Letting Jest transform behavior define production ESM behavior.
- Changing tsconfig, ESLint, or Jest prose without byte-consistent templates.
- Claiming reference-package proof without running the applicable package gate.

## Verification And Proof

- All internal Markdown links resolve.
- AGENTS, maintainer reference, recipes, checklists, and templates agree.
- Directory indexes enumerate every recipe, checklist, template, and decision artifact.
- JSON, YAML, JavaScript, Docker, Make, and TypeScript template syntax is checked with applicable local tools.
- Tooling claims cite current official documentation and exact pins are recorded in [templates/README.md](templates/README.md).
- `npm run verify` is green in every affected `typescript/reference/*` package.
- An invariant change links an accepted ADR.
- The PR follows [foundations/git-workflow.md](foundations/git-workflow.md).

Related: [onboarding-and-handoff.md](onboarding-and-handoff.md) for ownership changes and [templates/project-contributing.md](templates/project-contributing.md) for downstream repositories.
