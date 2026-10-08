# Temporal Agent Harness probe — durable approval vs current authority

Status: source-reviewed probe contract, not yet executed against the Temporal test server
Target: temporal-community/temporal-agent-harness
Source snapshot reviewed: c0f148f90499728a1fd305d7229113970c305f1a

## Scope correction

An earlier version of this probe assumed a stable workflow-side window between:

```
approval resolved
-> policy becomes stricter
-> activity dispatch
```

A deeper source pass does **not** support that assumption.

For an activity-backed tool, current main is effectively:

```python
await _apply_approval_policy(...)
return await workflow.execute_activity(...)
```

Once the parked approval gate resumes, the activity scheduling command follows directly on the same workflow execution path.

Therefore this probe no longer asks whether a normal runtime policy update can interleave before activity scheduling. The meaningful remaining boundary is later:

> an invocation was approved and durably scheduled, but its irreversible external effect has not happened yet; authority/policy is then revoked or tightened.

That is an execution-revocation question, not a pre-dispatch approval race.

## Current source contract already established

### Historical approval

A pending gate can become durably APPROVED and its `ToolApprovalResolved` event is retained in history.

### Runtime policy updates

`_apply_policy_update()` re-evaluates still-PENDING approvals.

A more restrictive update leaves pending calls pending.

Resolved approvals are not re-opened.

### Activity scheduling

After `_apply_approval_policy()` returns approved, the activity dispatcher immediately issues `workflow.execute_activity(...)`.

So the current architecture strongly suggests this semantic family:

```
approval for exact invocation
-> durable activity scheduling
```

with later policy updates primarily governing pending/future work.

The remaining question is whether revocation has any authority over already-scheduled but not-yet-effective work.

## Probe TAH-DAC-R1 — revoke after schedule, before irreversible effect

### Goal

Establish whether an already-authorized invocation remains authoritative once scheduled, even if a stricter policy is installed before its real external side effect.

This is a behavioral probe, not a proposed fix.

### Test shape

Add an activity-backed probe tool whose activity body has two phases:

1. publish/record that the Temporal Activity has started;
2. wait on a test-controlled barrier **before** incrementing the irreversible side-effect counter.

Then:

1. start the agent under `always_require_human_approval()`;
2. invoke the gated activity tool;
3. observe `tool_approval_requested`;
4. approve the exact `tool_id`;
5. observe `tool_approval_resolved(approved=True)`;
6. wait until the activity body has started and is blocked before the side effect;
7. install a stricter live policy / authority posture;
8. release the activity-side barrier;
9. observe whether the irreversible counter changes.

### Interpretations

If the side effect proceeds:

```
contract A:
approval authorizes the exact invocation durably;
later policy changes govern pending/future work,
not already-scheduled execution
```

If the activity is cancelled/prevented:

```
contract B:
current authority can revoke already-scheduled
but not-yet-effective execution
```

Either can be coherent. The important thing is to make the contract explicit.

## Probe TAH-DAC-R2 — worker restart while activity is authorized but not yet effective

Repeat R1, but restart the worker while the activity is blocked before its external effect.

Record:

- whether the activity is retried/resumed;
- whether historical approval is reused;
- whether latest policy state affects the retried activity;
- whether the external side effect occurs once, zero times, or more than once.

This combines durable approval with Temporal Activity retry semantics.

## Probe TAH-DAC-R3 — control: new call after restrictive posture

After the stricter policy is installed, submit a fresh equivalent tool call.

Expected under current documented behavior:

- the new/pending call is governed by the new policy;
- any remembered allow-list removed by the posture stays removed.

This distinguishes:
- already-authorized scheduled work;
- new work after the authority revision.

## Evidence to capture

- ordered AgentEvent stream;
- exact tool_id;
- ToolApprovalResolved;
- ToolStart/ToolEnd/ToolError;
- live approval policy before and after change;
- activity-start barrier state;
- external side-effect counter;
- worker restart point for R2.

## What this probe does not claim

It does not claim:
- a security vulnerability;
- that policy updates should cancel in-flight activities;
- that approval must be revalidated twice;
- that Temporal should introduce an authority revision abstraction.

It only asks whether the execution-revocation semantics are explicit and test-pinned.

## Durable Authority distinction

The corrected boundary is now:

```
historical approval
!=
necessarily revocable scheduled execution
```

and the system must choose/document whether:

```
approval -> exact invocation authority is durable
```

or:

```
current authority can cancel not-yet-effective work
```

This is more precise than the earlier pre-dispatch framing.

## Five-gate score after correction

External dependence:
High only if maintainers/users need revocation semantics for long-running privileged activities.

6-12 month credential:
High if the contract becomes a regression/documented behavior.

5-10 year portability:
High; scheduled-work revocation exists in workflows, job systems, payment pipelines, robots, and cloud control planes.

Anti-commoditization:
High; stronger models do not remove the need to define whether revoked authority stops already-scheduled effects.

Innovation + production consequence:
Potentially high, but only if the runtime experiment shows a meaningful policy/side-effect boundary.

## Stop condition

If maintainers explicitly define approvals as irrevocable authority for the exact invocation once scheduled, and that behavior is already sufficiently documented/tested, stop investing here.

The goal is to clarify a real execution boundary, not to manufacture an extra authorization layer.
