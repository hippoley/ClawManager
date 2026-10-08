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
