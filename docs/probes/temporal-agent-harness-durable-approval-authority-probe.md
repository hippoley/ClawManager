# Temporal Agent Harness probe — durable approval vs current authority

Status: source-reviewed probe contract, not yet executed
Target: temporal-community/temporal-agent-harness
Source snapshot reviewed: c0f148f90499728a1fd305d7229113970c305f1a

## Why probe this boundary

The harness deliberately makes approvals durable and supports runtime policy updates.

Current source also states that a more restrictive policy update:
- swaps the live policy;
- re-evaluates still-PENDING approvals;
- leaves already-resolved approvals unchanged.

The gate then wakes, finalizes the already-approved outcome, and returns to the tool dispatcher.

That behavior may be intentional.

The unresolved design question is:

> Does a historical approval represent immutable authorization for that exact tool call, or should a stronger current authority boundary be revalidated immediately before an external dispatch?

This probe should establish current behavior before proposing any change.

## Existing source behavior

### Resolution

`_apply_approval_policy(...)`:
- registers a pending approval;
- waits for human / auto / policy cascade;
- after an approved outcome, returns.

### Policy update

`_apply_policy_update(new_policy)` explicitly documents:

> A more restrictive update simply leaves pending calls pending.

It only iterates `pending_approval_entries()`.

Resolved approvals are retained as history but are not re-evaluated.

### Dispatch

The activity/workflow tool wrapper calls the approval gate before tool execution.

Therefore there is a narrow scheduling window to test:

```
approval resolves APPROVED
        ↓
policy becomes more restrictive
        ↓
tool has not externally executed yet
        ↓
does dispatch still proceed?
```

## Probe TAH-DAC-B1 — resolved approval followed by restrictive policy revision

### Goal

Establish current semantics without asserting that either result is wrong.

### Setup

Use an activity-backed side-effect probe tool whose activity body:
- increments a durable/local counter;
- records a `tool_start`;
- can be held behind a deterministic workflow-side barrier before actual activity scheduling if the existing test harness exposes such a seam.

Agent begins with:

```
ToolApprovalPolicy.always_require_human_approval()
```

### Sequence

1. Start one gated activity tool call.
2. Observe `tool_approval_requested`.
3. Submit explicit human approval for that exact `tool_id`.
4. Confirm `tool_approval_resolved(approved=True)` is in history.
5. Before the real activity is allowed to execute, update the live policy/posture to a stricter state that would not authorize a fresh equivalent call.
6. Release the dispatch barrier.
7. Observe whether the tool activity executes.

### Record, do not pre-judge

If activity executes:

```
current contract =
approval is authority for that already-approved invocation
even after later policy tightening
```

If activity does not execute:

```
current contract =
historical approval is preserved
but dispatch authority is revalidated
```

Either result is valuable because the contract becomes explicit.

## Probe TAH-DAC-B2 — crash/replay between approval and dispatch

Repeat the same logical sequence, but force a worker restart / workflow replay after approval and before external dispatch.

Record:
- whether approval is replayed without duplicate prompt;
- whether the latest live policy/posture is restored;
- whether execution occurs;
- which policy revision is visible in the event/audit record.

This is the durable-authority version of B1.

## Probe TAH-DAC-B3 — remembered approval vs later restrictive posture

1. Approve a tool with `remember=True`.
2. Confirm live policy allow-lists the tool.
3. Replace posture with a stricter policy that removes that allow-list.
4. Issue a *new* call of the same tool.

Current docs indicate a posture replaces the allow-list, so the fresh call should be governed by the new policy.

This control distinguishes:
- future-call policy replacement, which appears specified;
- already-resolved invocation authority, which is the B1/B2 question.

## Required evidence

Capture:
- ordered AgentEvent stream;
- approval id / tool id;
- approval result;
- policy state before and after update;
- tool_start/tool_end presence;
- actual side-effect counter;
- replay/restart point for B2.

Do not infer execution from an event alone if the activity body can provide a direct side-effect observation.

## Why this is not merely a Temporal question

The same boundary exists wherever:
- approvals live longer than processes;
- policy/identity can change during a wait;
- external side effects happen after durable recovery.

Potential systems include:
- durable agent runtimes;
- workflow engines;
- payment/refund agents;
- privileged automation;
- human-in-the-loop infrastructure;
- robot/IoT control planes.

## Possible future contract

Only if the probe demonstrates a real ambiguity worth changing should the runtime consider an explicit distinction such as:

```
approval_decision_revision
current_authority_revision
execution_generation
```

or a pre-dispatch authorization callback.

Do not add these abstractions unless a deterministic scenario proves they are needed.

## Five-gate score

External dependence:
High if the harness/Oso security direction needs an explicit replay-safe authority contract.

6-12 month credential:
High if converted into a merged regression or documented contract.

5-10 year portability:
High; authorization continuity across durable execution is framework-independent.

Anti-commoditization:
High; stronger models do not eliminate changing permissions or real side effects.

Innovation + production consequence:
Potentially high, but only after the probe distinguishes a real contract gap.
