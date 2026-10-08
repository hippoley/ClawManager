# Portable runtime-correctness precedents

Last refreshed: 2026-10-08

Purpose: preserve the durable engineering claims behind several OpenClaw contributions so they remain useful after the surrounding implementation, file names, and issue numbers change.

This is not a claim of maintainer status. It is a portability map: what was actually demonstrated, what higher-level invariant survives code churn, and what future role each precedent can legitimately support.

## Precedent A — custody is not authority

Primary upstream record:
- https://github.com/openclaw/openclaw/pull/166503

Durable fact:
A resource can settle or release physical custody while the logical operation still retains authority under its exact lifecycle owner.

Portable invariant:

> Custody settlement is not authority revocation.

Why it survives implementation churn:
The distinction applies anywhere a runtime separates a physical handle, lease, reader, connection, or registration from the lifecycle instance that owns permission to continue.

Future domains:
- hot reload
- worker generations
- database/session readers
- leases and capability handles
- distributed ownership transfer
- plugin/runtime replacement

Identity value:
Supports source-level correctness review where a system incorrectly equates "I no longer hold this object" with "this operation is no longer authorized."

---

## Precedent B — observable completion is not authoritative completion

Primary upstream records:
- https://github.com/openclaw/openclaw/issues/166466
- https://github.com/openclaw/openclaw/pull/166495

Durable fact:
A final-looking artifact can exist while publication is still incomplete. Consumers must not infer a stronger completion state than the publication owner has established.

Portable invariant:

> Observable final-looking state is not authoritative completion.

Why it survives implementation churn:
The same mistake occurs with files, checkpoints, manifests, queue rows, snapshots, delivery receipts, and replicated state.

Future domains:
- durable workflows
- checkpoint publication
- deployment/update systems
- artifact stores
- recovery protocols
- replication and reconciliation

Identity value:
Supports review of systems where downstream consumers infer completion from surface evidence instead of the authoritative publication/commit boundary.

---

## Precedent C — proof must cross the real ownership boundary

Primary upstream record:
- https://github.com/openclaw/openclaw/pull/154728

External reuse:
The PR author adopted two proof seams proposed by hippoley and explicitly credited the guidance.

Durable fact:
A helper-level regression can validate local logic while still failing to prove the behavior users observe through the production transport/owner path.

Portable invariant:

> Proof should cross the same ownership and transport boundary as the behavior being claimed.

Why it survives implementation churn:
The exact fixture names will disappear; the evidence discipline does not.

Future domains:
- RPC and transport correctness
- SDK conformance
- distributed worker tests
- security boundary testing
- failover/retry validation
- observability and replay

Identity value:
Supports reviewer roles where the contribution is not merely finding bugs but identifying the minimum production-faithful proof that resolves merge uncertainty.

---

## Precedent D — semantic progress follows validated execution

Current upstream position:
- https://github.com/openclaw/openclaw/issues/166770

Current public artifacts:
- openclaw-166770-semantic-progress-from-settled-tool-execution.md
- openclaw-166770-candidate-implementation-boundary.md

Durable fact under review:
A later response shape such as truncation must not erase or redefine successfully settled work. Conversely, response content alone must not be promoted into evidence that work executed.

Portable invariant:

> Semantic progress follows validated execution, not the enclosing response's terminal shape.

Supporting repository-native design:
OpenClaw already uses exact event-object identity plus owner generation for model-request lifecycle provenance. The candidate tool-execution repair reuses that pattern rather than creating a parallel lifecycle system.

Future domains:
- agent tool runtimes
- long-running workflows
- retry safety
- watchdogs
- streaming and partial results
- exactly-once / at-least-once side-effect handling
- distributed execution outcome reconciliation

Identity value if landed:
Would connect source diagnosis, exact-owner provenance, terminal truth, and watchdog/retry semantics into one production precedent.

Status:
Not yet externally adopted or merged. Do not cite as a landed credential.

---

# Cross-precedent structure

These records are not four unrelated bugs. They share a durable correctness pattern:

```
surface observation
    !=
authoritative runtime fact
```

Examples:

```
custody released
    != authority revoked

final-looking artifact exists
    != publication completed

helper-level behavior passes
    != production boundary proven

response ended with length
    != completed work did not happen
```

The reusable review question is:

> What fact is actually authoritative here, who owns it, and what evidence proves that fact at the same boundary?

This is narrower and more durable than branding every issue as a "lifecycle bug."

---

# Long-horizon credential ladder

## Already evidenced

1. Source-level semantic distinction.
2. Independent reviewer attribution.
3. Maintainer adoption of a canonical problem definition.
4. Upstream merged precedents.
5. Independent contributor reuse of proof guidance with explicit credit.

## Not yet evidenced

6. A hippoley-authored upstream correctness PR carried personally through review and merge.
7. Repeated maintainer routing to @hippoley for this class of problem.
8. A shared contract/test/conformance suite that downstream implementations must pass.
9. Cross-repository reuse of the same precedent without hippoley prompting it.
10. Formal maintainer, reviewer, working-group, or standards responsibility.

The next durable move should close one of these missing layers rather than adding more examples to layers 1-5.

---

# 6-12 month portability test

A precedent is worth keeping only if, after file names and implementation details change, it can still answer at least one of:

- What subtle runtime distinction did this person identify?
- Did an independent reviewer validate it?
- Did upstream production history change because of it?
- Did another contributor rely on the person's proof or repair reasoning?
- Does a regression or merged commit preserve the lesson?

The first three precedents currently pass this test.

---

# 5-10 year portability test

For a record to remain useful over 5-10 years, the value cannot depend on OpenClaw itself remaining important.

The portable asset must be one of:

- a durable distributed-systems invariant;
- an execution/settlement/retry principle;
- a proof methodology;
- a recognized reviewer/owner role;
- a reusable conformance contract;
- evidence of repeated judgment quality across independent systems.

The first four precedents currently support the first three categories. They do **not yet** establish the final three categories.

Therefore the long-horizon objective is not "collect more OpenClaw issues." It is:

```
repeat the same quality of judgment
across enough independent boundaries
that the person becomes the portable credential,
not the repository name
```

---

# Non-overclaim rule

Do not describe these records as:
- maintainer authority;
- formal standards work;
- repository-wide dependency on hippoley;
- proof that #166770 is fixed;
- evidence that future maintainers will automatically route work to hippoley.

Describe only what the public history proves.

That restraint is part of the credential: authoritative claims should follow authoritative evidence.


---

## Precedent E — yielded continuation custody is ownership transfer

Current upstream position:
- https://github.com/openclaw/openclaw/issues/166771

Public repair contract:
- `openclaw-166771-yielded-continuation-custody-transfer.md`

Durable fact under review:
A requester can relinquish active execution while still owing a future visible continuation through an exact durable child owner. Recovery must distinguish retained continuation custody from a permanently blocking recovery fence.

Portable invariant:

> **A yielded run may relinquish active execution while retaining a durable obligation; recovery must transfer that obligation to the exact continuation owner instead of confusing retained custody with permanent admission denial.**

Why it survives model progress:
The failure occurs in local admission/recovery ownership before a new model request begins. Better language models cannot repair a runtime that refuses to admit the continuation.

Future domains:
- durable agent orchestration
- workflow engines
- distributed sagas
- child-task joins
- crash/restart recovery
- actor/message ownership transfer
- long-running AI agents

Innovation / landing value:
The current source already exposes a precise continuation predicate (exact owner object + stable rearm generation), so the research question is not speculative. It can be converted into a bounded admission/recovery contract with positive settle-wake/user-input regressions and stale-owner negative controls.

Status:
Not externally adopted or merged yet. Do not cite as a landed precedent.

---

# Five-gate research selection rule

A new direction is worth public investment only when it can plausibly pass all five gates.

## Gate 1 — external dependence

Can a maintainer, contributor, downstream system, test suite, or policy actually consume the result?

Reject work whose only output is a clever explanation that nobody needs to depend on.

## Gate 2 — 6-12 month credential durability

Will public history still prove a scarce capability after implementation details move?

Prefer merged precedents, regression tests, explicit third-party credit, and canonical ownership records.

## Gate 3 — 5-10 year portability

Does the contribution encode a durable systems principle rather than knowledge of one repository's current file layout?

Prefer ownership, settlement, replay, recovery, authority, provenance, and conformance boundaries.

## Gate 4 — anti-commoditization

Would a much stronger general-purpose model make the problem disappear?

Prefer work that depends on real runtime state, exact ownership, durable evidence, physical/external side effects, governance, or cross-system conformance.

A model can help discover and implement these repairs, but cannot make the underlying accountability boundary unnecessary.

## Gate 5 — innovation with production consequence

Does the work introduce a genuinely better distinction, contract, proof method, or boundary that changes real system behavior?

Reject novelty that has no operational consequence.

The strongest candidates are where a new semantic distinction enables a smaller, safer production repair.

---

# 知行合一 / knowledge-action test

For this program, an idea is not considered learned until it changes observable engineering behavior.

```
知:
identify the authoritative fact and its owner

行:
encode it as a bounded repair, regression, provenance record, or upstream review intervention

验:
observe independent review/adoption/merge/reuse

复:
carry the same invariant into a new boundary without copying repository-specific wording
```

A concept that never reaches `行` is only interpretation.
An implementation without `知` is local patching.
A merged change without `验` is useful but not yet social capital.
Repeated independent reuse is what turns the result into position.


---

## Cross-repo convergence checkpoint

Independent LangGraph issues now provide evidence that the same runtime-accountability boundaries recur outside OpenClaw.

See:
- `runtime-accountability-cross-repo-convergence.md`

Important distinction:

```
cross-repo convergence
!=
cross-repo reuse
```

The existence of similar failures elsewhere validates the research direction but does not yet prove that another project depends on hippoley's work.

This changes the next-position criterion:

> Do not optimize for a fifth OpenClaw example if an opportunity exists to create a reusable contract, proof method, or repair that a second independent runtime can actually consume.

The highest-value next milestone is one of:

1. a second repository directly reuses a precedent/proof/contract;
2. a shared executable conformance suite emerges;
3. maintainers begin routing this problem class to hippoley;
4. a hippoley-authored production repair lands and later becomes precedent elsewhere.

That is the point where repository-specific credibility starts converting into portable technical authority.
