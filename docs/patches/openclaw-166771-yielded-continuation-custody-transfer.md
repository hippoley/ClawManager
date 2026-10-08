# OpenClaw #166771 — yielded continuation custody transfer

Source-reviewed against current main on 2026-10-08.

Status: repair/proof contract only; not executed and not a claim that the issue is fixed.

## Core distinction

A yielded requester is not terminally abandoned work.

When a requester yields to an admitted child batch, the requester still owes a future visible continuation, but execution custody has moved to the durable child continuation / settle-wake path.

Therefore:

> **Yielded continuation custody is a transfer of future-work ownership, not evidence that ordinary admission should remain permanently fenced.**

The reported state is contradictory only because two subsystems interpret the same retained fence differently:

- recovery admission treats the nonterminal `restartRecoveryRuns` fence as authoritative blocking state;
- orphan/yield reconciliation recognizes that a durable yielded child continuation still owns future settlement.

Both views are locally reasonable. Together they wedge the session.

## Source facts

### 1. Yield lifecycle intentionally preserves the fence

`projectMainSessionRecoveryLifecycle()` returns early for yielded waiting state rather than retiring the recovery fence.

That preserves continuation debt; deleting the fence unconditionally would be wrong.

### 2. The durable yielded continuation is already identifiable

`captureYieldedMainSessionContinuation()` finds a specific requester continuation by requiring:

- matching requester session route;
- non-collect child;
- `expectsCompletionMessage === true`;
- no `requesterTurnRunId`;
- `requesterSettleWake.requesterYieldBatch === true`;
- stable `rearmGeneration`;
- the child run included in the wake batch.

It returns an `isCurrent()` closure that revalidates the exact captured owner object and wake generation.

This is strong evidence that current main already has a precise continuation-ownership predicate.

### 3. Admission still treats the retained nonterminal fence as blocking

The main-session recovery admission path blocks an entry with retained nonterminal recovery state even when that state represents the yielded continuation above.

So the missing operation is not "delete recovery state" but **transfer / recognize custody at the existing admission boundary**.

## Preferred repair contract

A correct repair should preserve all three facts simultaneously:

```
exact yielded continuation still current
+
same session / lifecycle generation
+
no conflicting live recovery owner
        ↓
continuation custody is recognized
        ↓
settle wake and later user work can enter normal admission
```

while:

```
changed continuation
stale generation
replacement session
live conflicting recovery owner
reset/tombstone
        ↓
remain fenced
```

The repair must not weaken generation/reset/stale-owner protections.

## Why simple fixes are unsafe

### Do not delete every retained recovery fence on yield

The fence represents real continuation debt and crash-recovery ownership.

### Do not treat every yielded state as ordinary idle

Only an exact currently owned durable continuation should receive the transfer.

### Do not increase wake retries

Retries repeatedly encounter the same contradictory ownership state and do not repair custody.

### Do not infer ownership from run/session IDs alone

`captureYieldedMainSessionContinuation()` already demonstrates the stronger pattern: exact owner object + stable wake generation must still be current.

## Regression shape

Use persisted state matching the issue's minimal entry and a real durable yielded child continuation.

### Positive 1 — settle wake

1. Persist yielded requester state with retained nonterminal fence.
2. Register the exact durable child continuation / settle wake.
3. Attempt requester-settle work through the real recovery-admission owner.
4. Assert admission proceeds instead of `SESSION_WORK_START_CHANGED`.
5. Preserve the exact continuation until its settlement path owns the transition.

### Positive 2 — later user input

With the same yielded continuation still valid:

1. persist a later user message;
2. admit ordinary requester work through the production admission path;
3. assert the session can start the next assistant turn rather than remaining wedged by the transferred fence.

### Negative controls

Admission must remain blocked when:

- the captured continuation owner object is gone;
- `rearmGeneration` changed;
- the session was replaced/reset;
- lifecycle generation is stale;
- another live recovery owner currently owns the aggregate;
- the child continuation is not the exact yielded batch owner.

## Existing test seams

Current main already contains:

- `main-session-restart-recovery.test.ts` with a regression named
  `preserves the yielded global requester owner in a shared store`;
- the real recovery-admission owner used by `agent-command-recovery-owner.ts`;
- `server.subagent-settle-replay.harness.test.ts` for requester settle/replay behavior.

The highest-value proof is to bridge these existing seams rather than unit-test a new helper in isolation.

## Portable invariant

> **A yielded run may relinquish active execution while retaining a durable obligation; recovery must transfer that obligation to the exact continuation owner instead of confusing retained custody with permanent admission denial.**

This is independent of model quality. A stronger model cannot repair a local runtime that refuses to admit the continuation before any model request begins.

## Innovation / non-commoditization value

This boundary becomes more important as agents become more autonomous:

- more long-lived delegated work;
- more yields and resumptions;
- more concurrent child continuations;
- more crashes/restarts during orchestration;
- more need for deterministic responsibility transfer.

The durable technical asset is not "finding a prompt bug"; it is formalizing and proving ownership transfer across asynchronous runtime boundaries.
