# Durable Authority Continuity Conformance v0.1

Status: exploratory cross-runtime conformance draft
Last refreshed: 2026-10-08

This draft defines a small set of executable behavioral contracts for autonomous runtimes that combine:
- durable execution,
- tool approvals,
- retries/resume,
- runtime policy changes,
- external side effects.

It does **not** claim a vulnerability in any specific framework.
Its purpose is to make a currently fragmented correctness boundary testable across runtimes.

## Problem statement

Durable agent systems increasingly preserve approvals, tool calls, workflow state, and execution history across crashes and retries.

That durability creates a subtle risk:

> A durable fact that an action was approved in the past is not automatically the same thing as current authority to dispatch, accept, or publish that action after recovery.

The runtime needs to keep at least four layers distinct:

```
logical operation identity
!= execution generation / attempt identity
!= current authority state
!= external-effect truth / settlement evidence
```

A fifth layer is often necessary:

```
approval / policy decision history
```

because "approval happened" is historical evidence, while "execution is still authorized now" is a present-state question.

---

# Core invariants

## DAC-1 — Durable approval is historical evidence, not necessarily perpetual dispatch authority

A recorded approval MUST remain auditable across replay/restart.

But a runtime MUST NOT assume that the mere existence of that historical approval automatically authorizes a new external dispatch if a stronger current policy/identity boundary has changed.

Conformance scenarios should distinguish:

```
approval persisted
!=
fresh dispatch authority guaranteed forever
```

This contract is especially relevant when:
- user/operator identity changed;
- credential was revoked;
- spending or safety policy tightened;
- session ownership changed;
- organization/tenant changed;
- the original approval was scoped to one execution generation.

## DAC-2 — UNKNOWN external outcome is non-authorizing

If an external call times out after dispatch, the runtime/application MUST NOT promote timeout into proof of non-execution.

```
timeout
=> UNKNOWN
UNKNOWN
=> reconcile
UNKNOWN
!= safe retry
```

A new dispatch requires either:
- provider idempotency that returns the prior committed result; or
- reconciliation proving ABSENT plus a current authority check.

## DAC-3 — ABSENT does not recreate stale authority

Even when reconciliation proves the prior side effect did not commit, the runtime MUST re-evaluate current execution authority before creating a fresh attempt.

```
old intent
+ proven ABSENT
!=
automatic permission to retry
```

This protects against policy, identity, credential, or operator changes during recovery.

## DAC-4 — Late result acceptance is generation-scoped

A late completion from execution generation A MAY be recorded for audit/accounting.

It MUST NOT mutate present state, satisfy the current result, or recreate authority after:
- A was superseded by B;
- the operation was cancelled/terminalized;
- the relevant policy/authority revision changed.

```
same logical operation
!=
same execution generation
!=
current authority to publish result
```

## DAC-5 — Yield transfers obligation custody; it does not erase the obligation

A runtime that supports yield/subagents/continuations MUST distinguish:
- active execution ended;
- future obligation still owned by an exact durable continuation.

A yielded owner MUST NOT permanently block later admission merely because a nonterminal fence remains, if the exact continuation owner is still authoritative and the runtime contract permits transfer.

## DAC-6 — Approval supersession must be auditable

When an automatic evaluator, human decision, policy cascade, or runtime policy update settles the same gate, the runtime SHOULD expose enough durable evidence to answer:
- which decision actually won;
- which rule/policy version applied;
- which competing evaluator/decision was superseded;
- which execution generation consumed the result.

The event log should not make an old evaluator verdict look like the authority that actually released the tool.

---

# Minimum executable conformance scenarios

## Scenario A — approval survives crash with unchanged authority

1. Tool call is gated.
2. Human approves.
3. Worker crashes before tool dispatch completes.
4. Workflow resumes.
5. Identity, policy revision, session owner, and operation identity are unchanged.
6. Tool proceeds without asking for redundant approval.

Expected:
- prior approval remains auditable;
- no duplicate approval prompt;
- exactly one authoritative dispatch path.

This is the baseline durable-approval property.

## Scenario B — approval survives history, but authority is revoked before dispatch

1. Tool call is approved.
2. Before actual external dispatch, operator policy becomes stricter or the caller identity loses authority.
3. Workflow resumes/replays.

Expected:
- approval event remains in history;
- new dispatch is denied or requires fresh authorization according to policy;
- historical approval is not deleted or rewritten;
- no external side effect occurs.

This scenario deliberately distinguishes audit durability from live authority.

## Scenario C — external commit, caller timeout, resume

1. External provider commits side effect.
2. Caller times out before receipt.
3. Runtime resumes.

Expected:
- outcome is UNKNOWN until reconciled;
- runtime does not blindly redispatch;
- COMMITTED evidence reconstructs/reuses prior result;
- no duplicate side effect.

## Scenario D — timeout without commit, authority revoked before recovery

1. External dispatch attempt times out.
2. Reconciliation proves ABSENT.
3. Before recovery, authority is revoked.
4. Resume occurs.

Expected:
- no second dispatch;
- logical operation identity may remain;
- execution authority is denied under current state.

## Scenario E — stale completion after successor generation

1. Attempt A starts.
2. Recovery supersedes A with B, or terminalizes A.
3. A completes late.
4. A's result arrives after B/current state is authoritative.

Expected:
- A may be logged/accounted;
- A cannot overwrite B/current result;
- A cannot recreate dispatch authority.

## Scenario F — yielded continuation resumes without weakening stale-owner fences

1. Requester yields to exact durable child continuation.
2. Active requester segment ends.
3. Child settle wake or later user input needs admission.
4. Exact continuation is still current.

Expected:
- continuation custody is recognized/transferred;
- valid wake/input can proceed;
- stale generation, replaced session, changed continuation, or conflicting live owner remains blocked.

---

# Evidence model

A runtime claiming conformance should expose enough evidence to reconstruct:

```
logical_operation_id
execution_generation
authority_revision_or_equivalent
approval_decision_id
approval/policy revision
external_effect_receipt_state
current/superseded terminal status
```

Exact field names are implementation-specific.

The conformance property is semantic, not syntactic.

---

# Current cross-runtime evidence

## OpenClaw

Public records already demonstrate recurring boundaries around:
- custody vs authority;
- publication evidence vs authoritative completion;
- response shape vs settled execution;
- yielded continuation custody vs permanent admission denial.

## LangGraph

Public issues independently demonstrate:
- timeout after external commit followed by duplicate replay (#9185);
- checkpoint/pending-write ordering changing replay vs re-execution (#8039);
- nested resume re-running completed side-effecting work (#9106).

These establish that retry/replay correctness is not framework-specific.

## Temporal Agent Harness

Current public design already includes:
- durable approvals;
- runtime approval-policy updates;
- policy cascades;
- automatic evaluator supersession;
- durable/replayable approval events;
- tool execution after approval;
- criteria-version tracking.

That makes it a strong candidate environment for testing a more explicit authority-continuity contract.

This draft does **not** assert that its current behavior violates DAC-1 through DAC-6.

---

# Anti-commoditization rationale

A stronger code model can generate these tests faster.

It cannot make the underlying questions disappear:

- Was the payment actually committed?
- Which identity currently authorizes the action?
- Did the approval apply to this generation or only historical intent?
- May a stale result still change present state?
- Who owns the continuation after yield?
- Which evidence is authoritative after retry/replay?

These are accountability questions over real state and side effects.

As agent autonomy increases, the number of such boundaries increases.

---

# Innovation criterion

This draft should evolve only when a scenario catches a real ambiguity or production failure.

Do not grow it into an abstract taxonomy for its own sake.

A new rule belongs here only if it:
1. separates two facts that a real runtime currently collapses;
2. changes implementation or review behavior;
3. admits a deterministic regression;
4. can be reused across at least two independent runtimes.

---

# Position strategy

The durable identity target is not "author of a conformance markdown file."

The target progression is:

```
draft scenario
-> reproduce against real runtime
-> runtime maintainer accepts the distinction
-> executable fixture lands
-> second runtime reuses the same scenario
-> conformance becomes a shared review/CI artifact
```

Only at the latter stages does this become institution-level technical authority.

## Current status

- cross-runtime problem family: evidenced;
- conformance draft: created;
- first framework adoption: not yet established;
- executable shared harness: not yet established;
- formal standards/governance role: not yet established.
