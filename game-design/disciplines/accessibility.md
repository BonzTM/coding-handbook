# Accessibility

The accessibility bar this handbook enforces: a non-negotiable minimum every project ships, tiered adoption of the industry guidelines above that minimum, and the rule that accessibility is designed in from concept, never patched in after content lock.

## Default Approach

Two sources anchor this doc. The **Game Accessibility Guidelines** (`gameaccessibilityguidelines.com`) are the working reference: a collaborative industry checklist tiered by reach, impact, and implementation cost, grouped by impairment category (motor, cognitive, visual, hearing, speech, general). **AbleGamers APX** (`ablegamers.org/apx/`) supplies the design stance: research-driven patterns framed around ensuring players can modify the experience to their needs, addressed during design rather than as a post-hoc patch (`gdcvault.com/play/1025719/Accessible-Player-Experiences-A-New`). Accessibility is a player-reach and design-quality concern owned by this doc; it is not a difficulty setting — challenge tuning is owned by [difficulty-and-balance.md](difficulty-and-balance.md), and accommodations are never gated behind a difficulty mode.

### Minimum Bar

The Game Accessibility Guidelines identify the four most commonly complained-about issues as remapping, text size, colorblindness, and subtitle presentation (`gameaccessibilityguidelines.com/why-and-how/`). This handbook treats those four as the shipping minimum for every project, regardless of genre or platform:

- **Full input remapping.** Every gameplay action is rebindable to any available input; partial remapping (some actions fixed) does not count. The input layer is architected for remapping from the first playable, because a hard-coded input scheme is expensive to unwind later.
- **Text size.** UI and subtitle text can be scaled up, and every screen still functions at the largest setting. Text is rendered by the UI system, never baked into image assets where a size option cannot reach it.
- **Colorblind-safe signaling.** No gameplay-critical information is conveyed by color alone; every color-coded signal carries a redundant channel — shape, icon, pattern, position, or label (`gameaccessibilityguidelines.com/full-list/`). Redundant signaling is the requirement; a global recolor filter bolted on later is not a substitute for it.
- **Subtitles.** All speech and meaning-bearing audio is subtitled, with presentation quality (readable size, contrast against the scene) held to the same bar as the rest of the UI. Presentation specifics route to the guideline list above.

Dropping any of these four requires a decision record per [../decisions/design-decision-records.md](../decisions/design-decision-records.md) stating what is dropped and why the project's constraints force it — "we ran out of time" is a scheduling failure, not a rationale.

### Tiered Guidelines

Above the minimum bar, adoption follows the Game Accessibility Guidelines tier structure — **basic / intermediate / advanced**, tiered by number of players reached, impact, and implementation cost (`gameaccessibilityguidelines.com/full-list/`):

- **Basic tier: adopt by default.** Basic guidelines are low-cost and high-reach by construction; each one the project does not adopt is listed by name with a reason in the adoption sheet, not silently skipped.
- **Intermediate tier: decide per project at scoping.** Review the intermediate list during [../checklists/concept-intake.md](../checklists/concept-intake.md) and commit the selected set as scope in [../operations/scoping-and-production.md](../operations/scoping-and-production.md), so accessibility work is planned and costed like any other feature work.
- **Advanced tier: choose deliberately against the pillars.** Advanced guidelines are picked where they serve the project's [../foundations/design-pillars-and-vision.md](../foundations/design-pillars-and-vision.md) and audience, and tracked as backlog with an owner.
- **Adoption is a committed artifact.** Keep a per-guideline sheet (adopted / deferred-with-reason / not-applicable-with-reason) in the design docs, reviewed as part of [../checklists/design-review.md](../checklists/design-review.md). An adoption decision that lives in someone's head does not exist.
- **Console targets add the platform-holder reference.** Microsoft's Xbox Accessibility Guidelines (`learn.microsoft.com/en-us/gaming/accessibility/guidelines`) are the complementary checklist when shipping on or alongside Xbox platforms.

### Design-Time Not Patch-Time

The APX stance is the rule: accessibility is addressed during design, not retrofitted (`ablegamers.org/apx/`). Concretely:

- **Every new mechanic states its accommodation surface at design time.** The mechanic's design note answers: which senses and inputs does this demand, what is the redundant signal channel, and what can the player modify. A mechanic whose only feedback is a color flash or an audio sting is redesigned before it ships, not annotated for a future patch.
- **Accessibility enters at concept intake.** [../checklists/concept-intake.md](../checklists/concept-intake.md) asks the accessibility questions before production commits, because the expensive accommodations — remappable input architecture, scalable UI — are near-free when built in and structural rework when bolted on.
- **Modifiability is the design pattern.** Per APX, the goal is players modifying the experience to their needs: separate volume channels, toggleable effects, adjustable timing windows where the design allows. Options that reduce screen shake, flashes, and other high-intensity effects pair with the readability rule in [../foundations/game-feel.md](../foundations/game-feel.md) — juice must never destroy game-state legibility.
- **Accessibility options are UI like any other UI.** Option discoverability, defaults, and menu placement are held to the standards in [ux-and-onboarding.md](ux-and-onboarding.md); an accommodation the player cannot find before the first challenge fails at its job.

### Testing With Assistive Contexts

- **Test the options, not just their existence.** A shipped toggle that was never played with is untested code: run sessions fully remapped, at maximum text size, with subtitles on and audio muted, and with color-critical scenes checked under colorblind simulation. Each minimum-bar feature gets a full-session pass, not a menu screenshot.
- **Cheap internal proxies run every milestone.** A muted-audio run and a grayscale/simulation run flush out single-channel signaling violations without special recruiting. Proxies find regressions; they do not certify the experience.
- **Recruit players with disabilities into the playtest pipeline.** APX patterns come from research with disabled players; direct sessions through the standard protocol in [../quality/playtesting.md](../quality/playtesting.md) are the evidence that accommodations work, and observed failure to complete a task with an accommodation enabled is a defect, not feedback to triage away.
- **Regression-test accommodations like features.** Remapping, text scale, and subtitle rendering enter the ongoing test surface so a UI rework cannot silently break them; verify against [../checklists/playtest-readiness.md](../checklists/playtest-readiness.md) before each external test.

## Common Mistakes And Forbidden Patterns

- Treating difficulty modes as the accessibility plan — easy mode tunes challenge for players who can already execute the inputs; it accommodates nothing ([difficulty-and-balance.md](difficulty-and-balance.md)).
- Deferring accessibility to a post-launch patch, then discovering the input system and UI were never architected for remapping or scaling.
- Shipping partial remapping: a handful of preset layouts, or "rebind everything except these actions."
- Gameplay-critical information carried by color alone or sound alone, with a colorblind filter or a patch note offered as the fix instead of a redundant signal channel.
- Text baked into textures or art assets, unreachable by the text-size option and unscalable without an art pass.
- Subtitles that exist but fail presentation: unreadable size, no scene contrast, or missing meaning-bearing non-speech audio.
- An accessibility "pass" that opens each menu option once but never plays a full session with the option on.
- Skipping basic-tier guidelines silently, with no adoption sheet a reviewer can audit.
- Accessibility options buried where a player meets the first challenge before finding them — a [ux-and-onboarding.md](ux-and-onboarding.md) failure with an accessibility blast radius.
- New mechanics approved in design review with no stated accommodation surface, pushing the cost onto a future retrofit.

## Verification And Proof

- Minimum bar demonstrated in a build: every gameplay action rebinds, every screen functions at maximum text size, subtitles cover all speech and meaning-bearing audio, and no critical signal fails under colorblind simulation.
- The per-guideline adoption sheet exists and is current: every basic-tier guideline marked adopted or carries a named reason, intermediate selections match the committed scope in [../operations/scoping-and-production.md](../operations/scoping-and-production.md).
- Any dropped minimum-bar item has a decision record on file per [../decisions/design-decision-records.md](../decisions/design-decision-records.md).
- Milestone proxy runs on record: a muted-audio session and a grayscale/simulation session completed without loss of gameplay-critical information.
- Playtest evidence includes assistive-context sessions — full sessions played with each minimum-bar accommodation enabled, and sessions with disabled players logged through [../quality/playtesting.md](../quality/playtesting.md) reports.
- Design reviews of new mechanics show the accommodation-surface question answered ([../checklists/design-review.md](../checklists/design-review.md)).
- Accommodation regressions are caught before external tests: remapping, text scale, and subtitle checks appear in [../checklists/playtest-readiness.md](../checklists/playtest-readiness.md) sign-off.
