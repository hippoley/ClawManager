# Runtime Accountability — node map and entry criteria

Last refreshed: 2026-10-08

This document defines where to invest public contribution effort next.

North star:

> Move from repository-specific bug analysis to reusable runtime-accountability contracts that other runtimes, reviewers, and security/durable-execution teams can consume.

## Emerging node 1 — Durable Authority Continuity

Status: early / forming

External signals:
- Temporal is explicitly moving agent security, identity propagation, and policy enforcement into the durable execution platform.
- Temporal Agent Harness already combines durable approvals, runtime policy updates, supersession, replayable approval events, and tool dispatch.
- Temporal community discussions independently identify authorization/execution drift across long waits and retries.
- OpenClaw and LangGraph already provide adjacent evidence around stale authority, replay, and unknown external outcomes.

Core boundary:

```
durable approval history
!=
perpetual current dispatch authority
```

Entry condition:
- a deterministic probe proves an ambiguity or undesirable behavior at approval -> dispatch across policy/identity revision;
- or a maintainer asks for an explicit contract/test.

Current artifact:
- `durable-authority-continuity-conformance-v0.1.md`
- machine-readable vectors
- Temporal Agent Harness durable-approval/current-authority probe

Do not:
- file a security bug before behavior is reproduced;
- claim Temporal currently violates the contract;
- introduce an authority-revision abstraction without a failing scenario.

Why this node is attractive:
- anti-commoditization: very high;
- production consequence: very high;
- standards/conformance potential: high;
- early timing advantage: high.

---

## Emerging node 2 — External Effect Truth / Unknown Outcome Reconciliation

Status: independently recurring across runtimes

External signals:
- LangGraph #9185: provider commits payment, caller times out, resume duplicates the side effect.
- Temporal community discussion: uncertain external side effects should remain UNKNOWN and deny blind retry pending evidence.
- OpenClaw examples: missing receipt / terminal observation cannot be promoted into non-execution.

Core boundary:

```
error / timeout observed
!=
proof of external non-execution
```

Desired contract:

```
COMMITTED -> reuse/reconstruct prior result
UNKNOWN   -> reconcile only
ABSENT    -> re-check current authority, then maybe dispatch
```

Entry condition:
- a second runtime can consume a shared test vector;
- or an SDK/framework exposes a concrete reconciliation hook / receipt-state API.

Avoid:
- pretending generic idempotency alone solves every provider;
- claiming exactly-once across transaction domains without provider support.

Long-term destination:
- receipt/effect-state protocol;
- reusable conformance tests;
- integrations with payments, messaging, cloud control, robots/IoT.

---

## Emerging node 3 — Generation-Scoped Result Acceptance

Status: under-formalized

Core boundary:

```
same logical operation
!=
same execution generation
!=
current authority for a result to mutate present state
```

Why it matters:
- retries/replays can produce late results from superseded attempts;
- accounting may need to retain the physical result/cost while state acceptance must reject it.

Entry condition:
- reproducible late-result-after-supersession case;
- a runtime with explicit execution generations and observable stale completion.

Potential conformance:
- stale result can be logged;
- cannot overwrite current result;
- cannot revive execution authority;
- physical side-effect/accounting remains visible.

---

## Emerging node 4 — Continuation Custody Transfer

Status: concrete in OpenClaw, portable to workflow/actor systems

Core boundary:

```
active execution ended
!=
future obligation ended
```

Current anchor:
- OpenClaw #166771

Future bridge targets:
- nested workflows;
- agent subtask joins;
- actor handoff;
- sagas;
- human approval continuations.

Enter a second runtime only when an exact continuation owner is identifiable.

---

# Five-gate scoring

Score candidate contributions 0-2 on each gate.

1. External dependence
2. 6-12 month credential durability
3. 5-10 year portability
4. Anti-commoditization
5. Innovation + production consequence

Interpretation:

```
9-10  -> act publicly
7-8   -> probe privately / artifact first
5-6   -> observe
0-4   -> ignore
```

Current estimates:

| Node | Dependence | 6-12m | 5-10y | Anti-commodity | Innovation/landing | Total |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Durable Authority Continuity | 2 | 2 | 2 | 2 | 2 | 10 |
| Unknown Outcome Reconciliation | 2 | 2 | 2 | 2 | 2 | 10 |
| Generation-Scoped Result Acceptance | 1 | 2 | 2 | 2 | 2 | 9 |
| Continuation Custody Transfer | 1 | 2 | 2 | 2 | 2 | 9 |
| Generic agent debugging | 1 | 1 | 0 | 0 | 1 | 3 |
| Prompt/RAG glue | 1 | 1 | 0 | 0 | 1 | 3 |

---

# 知行合一 operating rule

For this research program:

```
知 -> identify a collapsed authority/truth boundary
行 -> encode it as a deterministic probe or conformance vector
验 -> obtain independent runtime behavior / maintainer response
立 -> land it as test, contract, review rule, or implementation
迁 -> reproduce reuse in a second system
制 -> become shared conformance/governance
```

Do not advance the narrative unless the preceding stage has evidence.

This keeps the work from becoming philosophy detached from runtime behavior.

---

# Exit conditions

Stop investing in a node if:
- no real runtime can reproduce the boundary;
- stronger existing guarantees make the distinction unnecessary;
- the only value is explanation rather than changed behavior;
- the problem disappears entirely with ordinary model improvement;
- there is no plausible maintainer/downstream consumer;
- another contributor has already occupied the exact implementation slot and the remaining work would be redundant.

The goal is sparse, durable position — not maximal activity.


---

## 2026-10-08 position update — Temporal Agent Harness #174

Public entry:
- https://github.com/temporal-community/temporal-agent-harness/issues/174

This moves **Durable Authority Continuity** from a private/public artifact-only node into a second runtime's own issue tracker.

Current stage:

```
知 -> complete enough to state the distinction
行 -> public probe/contract issue opened
验 -> waiting for external behavior
立 -> not yet
迁 -> started, not completed
制 -> not yet
```

Interpretation:

- This is stronger than independent convergence because the contract is now visible inside another runtime's design surface.
- It is weaker than cross-repo reuse because no maintainer or contributor has accepted, cited, requested, or implemented it yet.
- Do not add more comments unless new source evidence or an external response changes the decision boundary.

Next identity upgrade trigger:
- maintainer confirms one intended semantic and asks for a regression;
- contributor implements the probe/test;
- issue is incorporated into docs/contract/tests;
- or the same conformance vector is reused by another runtime.


---

## 2026-10-08 executable-model update

Durable Authority Continuity has now advanced one step beyond prose/vector design.

New executable artifact:
- `tools/durable_authority_reference_model.py`

Validation:
- 6/6 reference scenarios passed in local execution.
- Evidence recorded in `docs/evidence/durable-authority-reference-model-validation.md`.

Current stage:

```
知 -> stable enough to state the distinctions
行 -> spec + vectors + executable reference model + Temporal #174
验 -> internal reference model passed; external runtime behavior still pending
立 -> not yet
迁 -> cross-repo public entry started
制 -> not yet
```

Identity rule:

Do not treat the executable reference model as external adoption.
Its value is that a third-party runtime can now compare behavior against a concrete state machine instead of prose alone.

The next legitimate upgrade remains external:
- maintainer response on Temporal #174;
- a framework adapter reproducing one DAC scenario;
- a merged regression or documented contract;
- or reuse by a second runtime.


---

## 2026-10-08 fit correction — Temporal Harness vs durable authorization

A deeper source pass changes the node fit.

### Temporal Agent Harness #174

Issue:
- https://github.com/temporal-community/temporal-agent-harness/issues/174

What the source actually supports:

- `ToolApprovalPolicy` controls gating / auto-approval;
- restrictive updates re-evaluate still-pending calls;
- there is currently no general deny-list / authority-revision surface;
- after an approval gate returns, activity scheduling follows directly on the same workflow execution path.

Therefore:

```
ToolApprovalPolicy tightening
!=
general authorization revocation
```

and the harness is **adjacent evidence**, not yet the best production host for Durable Authority Continuity.

Public correction was posted to #174 rather than forcing an implementation thesis onto the wrong abstraction.

### Node status change

```
Temporal Agent Harness approval layer
FROM: primary Durable Authority implementation candidate
TO:   semantic clarification / adjacent precedent
```

Do not open a PR here unless maintainers identify an explicit revocation/cancellation contract they want to add.

---

## Promoted forming node — identity propagation + policy enforcement inside durable execution

Temporal publicly announced on 2026-10-06 that the Oso team joined to accelerate agent security, explicitly naming:
- identity propagation;
- policy enforcement;
- agents calling tools and acting on vital data;
- security built natively into the execution platform.

This is a much better conceptual host for Durable Authority Continuity than the current Harness approval gate.

Current public Temporal primitives already expose adjacent pieces:
- context propagation across Workflow / Activity / Child Workflow boundaries;
- interceptors for authorization/header manipulation;
- server-side ClaimMapper / Authorizer for Temporal API calls.

But these are not yet the same as an agent-action authority contract.

The forming gap is:

```
caller / user / agent identity
+
policy decision
+
durable workflow history
+
activity execution generation
+
external side effect
        ↓
what authority must be propagated / revalidated / revoked across recovery?
```

### Enter only when a real public surface appears

Act publicly when one of these appears:

1. an Oso/Temporal agent-security repository or SDK API for identity/policy propagation;
2. a design issue defining authorization across Workflow -> Activity / tool boundaries;
3. a concrete sample where user/agent authority is persisted or propagated into long-lived tool execution;
4. an API that distinguishes historical approval from current execution authorization;
5. a maintainer request for conformance / replay / revocation semantics.

Until then:

```
observe + preserve early evidence
!=
invent an API or issue before the owning surface exists
```

### Why this node outranks the Harness gate

It has a plausible path to:
- real enterprise dependence;
- security/governance ownership;
- cross-SDK semantics;
- durable identity contracts;
- standards/conformance;
- long-horizon role durability.

This is the current highest-upside early node for the Runtime Accountability program.


---

## 2026-10-08 promoted seam — Principal Attribution vs Execution Authorization

Public Temporal documentation issue:
- https://github.com/temporalio/documentation/issues/5451

Current Temporal primitives already separate three useful mechanisms:

```
API authorization
-> Authorizer / ClaimMapper

history attribution
-> Principal Attribution

application metadata propagation
-> Context Propagation / Interceptors
```

The new public clarification asks Temporal to make one non-equivalence explicit:

```
Principal Attribution
!= propagated application identity
!= current authorization for an external side effect
```

Why this is a better early node than pushing Agent Harness #174:

- it maps onto real existing Temporal primitives;
- it does not invent a revocation surface that the Harness does not have;
- it sits directly beside the announced agent-security / identity-propagation direction;
- it can mature naturally into runtime conformance if a true execution-authorization API appears.

Current evidence stage:

```
existing public primitives       = YES
public cross-repo contract issue = YES (#5451)
maintainer response              = NOT YET
docs clarification merged        = NOT YET
runtime authorization API        = NOT YET
cross-runtime conformance reuse  = NOT YET
```

### Enter deeper only on external signal

Escalate from docs clarification to implementation/conformance only if one of these occurs:

1. Temporal maintainers confirm the distinction and identify an owning API;
2. Oso/Temporal publishes an agent authorization/identity propagation SDK surface;
3. Principal/identity metadata becomes consumable at Activity/tool execution for enforcement;
4. a concrete revocation/replay bug demonstrates the gap;
5. another runtime reuses the same attribution/authority distinction.

Until then, do not create speculative authorization state inside the Harness.


---

## 2026-10-08 promoted implementation node — Open Agent Auth revocation continuity

Target: alibaba/open-agent-auth

New evidence:
- project docs promise AOAT revocation / immediate invalidation and per-use validation;
- TokenRevocationService exists with revoke(token) and isRevoked(token);
- current Resource Server AOAT acceptance path does not visibly consume revocation state;
- no RS-side introspection / remote revocation-state consumer was found;
- a private report was submitted through the repository-defined SECURITY.md channel;
- a cross-service conformance case is now captured in docs/specs/open-agent-auth-aoat-revocation-conformance.md.

Important correction:

The durable repair target is NOT merely "inject the in-memory revocation service into AoatValidator".
In split AS/RS deployments, authority state must propagate across service boundaries.

Portable contract:

authorization artifact historically valid
!=
current resource-server execution authority

and:

cacheable cryptographic validity
!=
cacheable revocation/current-authority state

This is the first independent implementation where Runtime Accountability has produced a concrete distributed authorization-continuity finding rather than only conceptual convergence.

Current stage:

source finding = YES
responsible disclosure = YES
conformance contract = YES
maintainer acknowledgment = NOT YET
maintainer confirmation = NOT YET
fix/regression landed = NOT YET
public attribution = NOT YET

Do not count this as cross-repo adoption until external confirmation exists.
