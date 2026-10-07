# OpenClaw Runtime Evidence & Authority Conformance Node

## Thesis

The next durable position is not "review more pull requests." It is to become useful at the layer
where OpenClaw turns implicit runtime semantics into executable conformance.

Three existing repository control surfaces are converging:

1. **Real behavior proof policy**
   - `scripts/github/real-behavior-proof-policy.mjs`
   - PR evidence policy, proof labels, AGENTS/testing guidance
2. **Runtime authority/custody semantics**
   - session/tool/message/delivery/native/placement authority owners
   - custody, admission, currentness, generation, retirement, settlement
3. **Reusable runtime contract suites**
   - PR #160649: ACP turn contract testing
   - request identity, event ordering, terminal authority, cancellation, stream closure

The strategic position is the intersection:

```text
Evidence binding
      ▲
      |
Runtime authority ─── Executable conformance
```

## Why this is now credible

### Canonical problem → maintainer implementation → merge

`hippoley` authored canonical issue #166466.

The issue was classified:

- `impact:data-loss`
- `clawsweeper:source-repro`
- `clawsweeper:fix-shape-clear`
- `clawsweeper:queueable-fix`
- diamond-lobster issue quality

Maintainer PR #166495 explicitly closed #166466 and was squash-merged as
`a86dd383534f8b3b6d1ebdc7e4b7601db65c6ad3`.

That establishes a complete public chain:

```text
problem definition
→ canonicalization
→ executable invariant
→ maintainer implementation
→ runtime qualification
→ merge
```

### Repeated evidence-binding failures

Recent upstream work repeatedly required distinguishing:

- shared-shard timing from per-file timing (#166361);
- handler invocation from real Gateway transport (#154728);
- local throwaway state from server-owned canonical state (#154728);
- fs-safe syscall behavior from OpenClaw Doctor entrypoint behavior (#166495);
- dependency/runtime proof from published-driver compatibility (#166495);
- raw multiline fixture input from production-normalized recall representation (#166514).

This is one recurring class:

> Evidence is useful only when it is bound to the exact claim, head, owner, boundary and outcome.

### Repeated authority/custody failures

Recent work also repeatedly exposed state conflation:

- native custody is not discovery authority (#166503);
- child traffic is not parent continuation (#166511);
- ordinary snapshot consistency is not prepared replay authority (#166487);
- final-manifest presence is not sealed recovery authority (#166466/#166495);
- stream closure is not terminal-result settlement (#160649 contract suite).

This is another recurring class:

> Retained custody, observation, presence or liveness must not be confused with authority.

## The institutional opening: PR #160649

PR #160649, `improve(acp): add reusable runtime turn contract suite`, is the clearest current
institutional node.

It converts implicit `AcpRuntime.startTurn` behavior into a private reusable conformance suite and
uses ACPX as the first adopter.

Current contract coverage:

- stable request identity;
- ordered event projection;
- authoritative completed/failed terminal results;
- terminal settlement independent from event-stream closure;
- prompt readiness;
- cancellation forwarding and cancelled terminal outcome;
- `closeStream` unblocks event consumption without fabricating terminal settlement.

This is the correct architectural direction because it moves semantic review from:

```text
reviewer remembers invariant
→ adapter implements its interpretation
```

to:

```text
shared contract
→ any adapter proves observable compatibility
```

The current suite explicitly does **not** cover process termination, crash cleanup, trust boundaries,
container readiness or per-user isolation. That boundary should be respected.

## Position to occupy

Do not attempt to become the owner of #160649 by commentary.

Instead become the person who repeatedly contributes the next executable invariants when real
failures demonstrate that the shared contract is incomplete.

The desired community association is:

> hippoley closes runtime-boundary, authority and evidence gaps, then turns repeated findings into
> reusable conformance.

That is stronger than "good reviewer" and more portable than subsystem-specific authorship.

## First reusable vocabulary

Prefer language already consistent with the repository:

### Evidence binding

A runtime proof should bind:

```text
case
head
owner
boundary
outcome
```

A proof that does not bind these facts is an observation, not conformance evidence.

### Authority separation

Keep distinctions explicit:

```text
custody != authority
observation != ownership
presence != provenance
stream closure != terminal settlement
snapshot consistency != replay authority
wall time != active execution time
```

These should become executable only after repeated concrete failures justify a shared check.

## What would justify moving from PR comments into shared policy/code

Escalate a repeated finding into a shared invariant only when at least one of these occurs:

1. the same failure class appears in three independent owners/adapters;
2. ClawSweeper repeatedly asks for the same proof boundary;
3. two implementations independently recreate the same lifecycle/authority check;
4. a maintainer-authored fix adopts language from an earlier canonical issue;
5. a reusable suite already exists and a new invariant fits its observable contract without forcing
   architecture.

At that point, prefer:

```text
shared contract test
> shared helper
> policy regression
> documentation
> another explanatory PR comment
```

## High-value nodes to watch

### 1. ACP runtime contract suite (#160649)

Best opening:
- second real adapter adoption;
- first concrete bug showing a missing cross-adapter turn invariant;
- terminal/cancellation/identity behavior repeated outside ACPX.

Do not broaden it prematurely into process/trust/container behavior.

### 2. Real behavior proof policy

Watch for repeated reviewer distinctions involving:

- exact-head evidence;
- canonical-owner state;
- real transport versus handler/mock boundary;
- published-driver/runtime composition;
- stale or foreign evidence.

A policy contribution becomes justified when the same missing binding is causing repeated external
PR friction, not merely because a richer taxonomy would be elegant.

### 3. Authority/custody owners

Watch for bugs crossing:

- message injection;
- pending input;
- session writer delivery;
- native process/subagent admission;
- placement turn authority;
- tool authority.

The opportunity is inconsistent invariant enforcement across owners, not adding a generic authority
framework for its own sake.

## Positioning discipline

Avoid:

- commenting on every relevant PR;
- proposing a large generic "conformance framework";
- asking for maintainer/collaborator status;
- creating terminology disconnected from repository vocabulary;
- treating every test failure as a policy problem.

Prefer:

```text
one repeated semantic distinction
→ one canonical issue or minimal patch
→ one accepted upstream implementation
→ one shared executable invariant
```

## Current position scorecard

### Proven

- evidence closer: #166361 merged after precise blocker evidence;
- runtime-boundary closer: #154728 author explicitly thanked `@hippoley` for the Gateway harness;
- canonical problem definition: #166466;
- maintainer implementation from that canonical issue: #166495;
- mainline landing: #166495 → `a86dd383...`.

### Emerging

- authority/custody semantic reviewer: #166503, #166511, #166487;
- proof-boundary reviewer: #154728, #161046, #166514;
- conformance-node candidate: connection to #160649 shared runtime contract testing.

### Not yet earned

- shared contract authorship;
- policy ownership;
- maintainer routing by default;
- cross-adapter conformance ownership;
- collaborator/maintainer role.

Those should remain outcomes, not requested titles.

## Next-position criterion

The next identity step is reached when one of the following becomes public and repeatable:

```text
maintainer routes a runtime-boundary case to hippoley
OR
a shared contract/policy test contains an invariant first surfaced through hippoley's upstream work
OR
a second canonical issue follows the #166466 pattern and lands through maintainer implementation
```

That is the next real admission ticket.


## New evidence that the node is becoming externally legible

PR #166503 now supplies the first explicit `invariant -> production code -> reviewer attribution`
chain.

The accepted distinction was:

```text
native reader custody != discovery authority
```

The current production head removed native-sequence presence from the post-cleanup authority
predicate while preserving genuine revocation/closing checks and the earlier physical-custody
filter.

ClawSweeper's formal evidence section explicitly links the `hippoley` handoff and states that the
repair agrees with “hippoley's useful predicate distinction.”

This matters because it crosses a new threshold:

```text
useful comment
< author acknowledgement
< production adoption
< reviewer attribution to the invariant
```

The desired next threshold is not another acknowledgement. It is reuse:

```text
same distinction
→ shared helper / contract case
→ applies across multiple owners
```

That is the point at which the position becomes institutional rather than personal.


## Authority consolidation watch: PR #166259

PR #166259 is a stronger structural signal than an isolated bug.

It moves conversation registration, binding mutations and delivery recovery into existing database
workers while preserving:

- transaction-local ownership facts;
- live revocation checks;
- final-effect authority around the concrete platform-method handoff;
- release before transport settlement;
- outcome-unknown semantics after accepted initiation rather than granting replay permission.

The current reviewer has no correctness finding on those authority semantics; the remaining blocker
is per-flow SQL overhead.

This makes #166259 a **watch node**, not a comment target.

The important future trigger is:

```text
if another bug appears because conversation binding / delivery / revocation owners disagree
→ do not patch only that call site
→ test whether a shared authority invariant/helper/contract is now justified
```

Together with #166503, this provides two distinct authority surfaces:

```text
session-reader custody / discovery authority
conversation binding / delivery authority
```

A third independent occurrence would be a strong threshold for shared institutionalization.
