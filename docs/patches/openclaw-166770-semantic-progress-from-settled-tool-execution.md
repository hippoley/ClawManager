# OpenClaw #166770 — semantic progress must follow settled tool execution

Source-reviewed against current default-branch files retrieved on 2026-10-08:
- `src/agents/embedded-agent-runner/run/attempt.model-diagnostic-observation.ts` blob `0082099b6affeff83f2b48d1dd1d30cf7cf17ee1`
- `src/logging/diagnostic-run-activity.ts` blob `0ab2033d8ee3f8ff924f57ea76287b38ffc48213`

This note is source-reviewed only. I have not executed the regression or a candidate patch.

## Problem boundary

Issue #166770 reports a response that successfully executed async tool calls, then ended with `stopReason: "length"`. The watchdog later treated the run as if no semantic progress had occurred.

The current model-result classifier only treats a normalized tool call as semantic when the enclosing result ends with `stopReason === "toolUse"`:

```ts
const hasExecutableToolCall =
  result.stopReason === "toolUse" && result.content.some(isNormalizedToolCall);
```

That is a useful model-result heuristic, but it is not authoritative evidence that tool work did or did not execute.

The stronger fact already exists elsewhere: trusted tool lifecycle events distinguish

```
tool.execution.started
tool.execution.completed
tool.execution.error
tool.execution.blocked
```

and the run-activity owner already tracks active tools.

So the semantic invariant should be:

> **Validated, successfully settled tool execution under the current run owner is semantic progress even if the enclosing model response later ends with `length`.**

The converse also matters:

> A normalized tool call, a tool start, a failed/blocked tool, or a delayed terminal event from a stale owner is not enough to clear repeated-request stagnation evidence.

## Why the obvious one-line fixes are unsafe

### Do not broaden `isSemanticModelCallResult()` to every `length` result containing tool calls

That would infer execution from model output shape. A tool call can be present without having successfully executed.

### Do not make every `recordToolEnded()` semantic

The async diagnostic listener currently groups `completed`, `error`, and `blocked` into the same `recordToolEnded()` path.

Turning that shared path into semantic progress would incorrectly treat failed and blocked tools as successful work.

It would also let delayed public terminal events refresh a replacement owner unless the exact-owner fence is preserved.

### Do not make every synchronous owned `phase: "end"` semantic

`markDiagnosticOwnedToolActivity()` currently receives only `phase: "start" | "end"`. The API does not encode whether an end represents successful completion versus terminal failure.

Without carrying terminal status, an owner-local `end` cannot safely become authoritative semantic progress.

## Narrow repair contract

The production change should publish semantic progress at the **successful tool-settlement boundary**, while preserving exact-owner validation.

A safe shape is one of:

1. Extend the exact-owner tool activity API so its terminal call carries a success/error terminal kind, and only successful settlement calls `touchSemanticSessionActivity(...)`; or
2. Bind trusted `tool.execution.completed` to the exact current diagnostic owner before promoting it to semantic progress.

Whichever implementation is smaller in the owning code, the required properties are the same:

```
accepted by current owner
    +
tool execution actually completed successfully
    +
terminal fact still belongs to that owner
        ↓
semantic progress
        ↓
clear repeated-request stagnation for that run
```

Do **not** use final model stop reason as the execution truth source.

## Regression contract

### Positive regression

Create an active diagnostic run owner with repeated-request stagnation evidence, then record:

1. tool start
2. successful tool completion for that exact owner
3. a later model result whose terminal stop reason is `length`

Assert that the successful tool settlement refreshes semantic progress / clears the run's repeated-request no-progress evidence before the later truncated model result is observed.

The same execution should remain progress regardless of whether the enclosing final response is `toolUse` or `length`.

### Negative regressions

Keep repeated-request evidence intact for:

- `tool.execution.error`
- `tool.execution.blocked`
- tool start without terminal success
- a terminal completion belonging to a closed/replaced owner
- a delayed unbound completion event that cannot prove current-owner identity

These negatives are important: the repair should preserve execution truth, not turn generic tool activity into progress.

## Suggested proof ladder

1. Focused run-activity regression: red on main, green on candidate.
2. Existing repeated-request/watchdog tests stay green.
3. Exact-owner negative cases above.
4. If practical, replay the issue's classifier shape with identical successful tool execution followed by `toolUse` vs `length` and show both retain the same semantic-progress age.

## Reusable invariant

> **Semantic progress follows validated execution, not the enclosing response's terminal shape.**

This keeps the watchdog aligned with authoritative work that actually settled, while preserving the existing safeguards against incomplete calls, failed tools, stale owners, and delayed diagnostics.
