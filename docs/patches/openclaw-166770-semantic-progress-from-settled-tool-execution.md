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


## Ownership seam audit

A second source pass narrowed the implementation boundary further.

### Public terminal events do not currently carry exact-owner provenance

`src/infra/diagnostic-tool-execution-liveness.ts` only attaches liveness metadata to
`tool.execution.started`. The terminal `tool.execution.completed/error/blocked` events do not
carry the diagnostic embedded-run owner generation.

That means the async listener in `diagnostic-run-activity.ts` cannot safely turn an arbitrary
queued `tool.execution.completed` event into semantic progress solely from event shape.

### The owner-local path is authoritative but loses terminal status

`markDiagnosticOwnedToolActivity(owner, ...)` validates that
`activeDiagnosticOwners.get(owner.generation)?.owner === owner`, so it has the exact ownership
property the repair needs. However, the API currently collapses terminal state to
`phase: "end"`.

The worker owner maps every non-update tool terminal to that `end` shape:

```ts
phase: event.payload.phase === "start" ? "start" : "end"
```

Therefore a safe repair needs to preserve **both** facts at one boundary:

```
exact current owner
+
terminal kind === successful completion
```

### Consequence for the production patch

The smallest trustworthy patch should extend the owner-local tool activity contract rather than
promoting the unbound async terminal listener.

For example, conceptually:

```ts
markDiagnosticOwnedToolActivity(owner, {
  toolName,
  toolCallId,
  phase: "start" | "completed" | "error" | "blocked",
  deadlineAtMs,
})
```

and only the `completed` branch should:

1. retire the active-tool marker;
2. call semantic progress for that exact run owner;
3. clear repeated-request stagnation evidence.

`error` and `blocked` should retire active work without claiming semantic progress.

The exact field names can follow the existing worker live-event terminal vocabulary; the important
part is not to invent a second ownership system or infer success from the later model response.

This is now the preferred patch direction over modifying `isSemanticModelCallResult()`.


## Exact regression seam on current main

Current main already has the right watchdog-level fixture in:

`src/logging/diagnostic.test.ts`

The existing test:

`does not recover repeated requests after semantic output resets the clock`

is the natural behavioral anchor for #166770. A focused sibling regression should preserve its setup and replace the semantic model-output reset with a **successful exact-owner tool settlement**.

A second useful owner-fencing fixture already exists later in the same file:

`prunes stale same-key activity while preserving a different owner's fresh tool`

That case proves the activity layer already distinguishes a fresh exact owner from stale same-key activity.

### Suggested red/green regression shape

1. Start diagnostics and a current embedded-run owner.
2. Seed repeated-request stagnation exactly as the existing semantic-output test does.
3. Start a tool under the current owner.
4. Complete it successfully through the candidate owner-local terminal API.
5. Advance beyond the previous repeated-request abort threshold.
6. Assert no `repeated_model_requests_without_progress` recovery is requested.
7. Observe a later model result ending in `length`; assert it does not undo the already-published semantic progress.

### Negative siblings

Using the same owner fixture, verify the stagnation clock is **not** cleared when the terminal kind is:

- error
- blocked

and is not cleared when:

- the owner has already been closed/replaced before the terminal signal;
- only an unbound queued public terminal event arrives for a stale run.

This keeps the proof at the watchdog boundary rather than only unit-testing a helper.

### Why this seam is preferable

It proves the user-visible policy directly:

```
successful work happened recently
+
later response truncation
=> watchdog must not claim no semantic progress
```

while the exact-owner fixture prevents the repair from becoming a generic "any tool event resets the clock" rule.


## Coverage correction: worker-local ownership is not sufficient

A later source audit found an important coverage gap in the earlier preferred patch direction.

`markDiagnosticOwnedToolActivity(owner, ...)` is currently used by the **worker turn owner** path. The canonical #166770 report, however, concerns native/embedded async tool execution. Therefore, extending only that worker-local API would preserve the right semantics but would **not** cover the reported embedded path.

This rules out a worker-only patch as the canonical fix.

### Embedded path facts

The embedded runner already owns an exact `DiagnosticEmbeddedRunOwner` in
`attempt-stream-prepare.ts`.

The ordinary tool wrapper emits trusted terminal events from
`agent-tools.before-tool-call.wrapper.ts` after classifying the actual result as:

- `tool.execution.completed`
- `tool.execution.error`
- `tool.execution.blocked`

But the hook/tool execution context currently carries run/session identity rather than the exact
diagnostic owner generation.

So the missing fact is not terminal truth; terminal truth already exists. The missing fact is
**terminal truth bound to the exact embedded-run owner**.

### Revised preferred repair boundary

The canonical repair should thread exact embedded-run diagnostic provenance into the ordinary tool
execution terminal accounting path, using the existing diagnostic owner rather than inventing
another generation system.

Conceptually:

```
DiagnosticEmbeddedRunOwner
        ↓
tool execution admission / terminal provenance
        ↓
tool.execution.completed
        ↓
verify exact current owner
        ↓
semantic progress
```

while `error`, `blocked`, closed-owner, replacement-owner, and unbound delayed terminal cases do
not clear repeated-request stagnation.

Implementation can choose the smallest private mechanism that preserves this property:

1. pass/capture a private diagnostic-owner provenance object alongside the wrapped tool execution;
2. attach that provenance to trusted tool lifecycle metadata and consume it in run-activity
   accounting; or
3. invoke an owner-bound terminal accounting helper directly from the wrapped execution path.

The important constraint is that **runId/sessionId equality is not a substitute for exact owner
provenance**, and the provenance should remain private to runtime diagnostics rather than becoming a
plugin-facing authority surface.

### What is now rejected

The following patch shapes are explicitly rejected as incomplete or unsafe:

- broadening `isSemanticModelCallResult()`;
- treating generic `recordToolEnded()` as semantic;
- extending only the worker-owner path;
- treating matching run/session IDs as proof of current ownership.

This correction narrows the actual canonical fix to the embedded execution owner boundary that
#166770 exercises.


## Existing provenance pattern to reuse

Current main already has the exact pattern needed for a safe implementation in
`diagnostic-model-request-provenance.ts` and `diagnostic-model-request.ts`.

Core model-request lifecycle binds:

```
exact event object identity
+
exact owner generation
+
phase
```

through a private `WeakMap<object, provenance>`, then exposes the provenance only through trusted
diagnostic metadata. The source explicitly states:

> Exact event and generation identity are core-only authority; payload fields cannot forge either.

That is a better implementation precedent for #166770 than inventing a new ownership mechanism.

### Preferred implementation family

Mirror the model-request pattern for core tool lifecycle:

```ts
type CoreToolExecutionLifecycleProvenance =
  | { generation: CoreModelRequestOwnerGeneration; phase: "started" }
  | { generation: CoreModelRequestOwnerGeneration; phase: "completed" }
  | { generation: CoreModelRequestOwnerGeneration; phase: "error" }
  | { generation: CoreModelRequestOwnerGeneration; phase: "blocked" };
```

The concrete type split can be smaller if start liveness remains separate. The invariant is what
matters:

- provenance is attached by core runtime code, not accepted from payload fields;
- terminal provenance carries the same exact owner generation as the admitted embedded run;
- run-activity consumes the metadata and verifies that generation is still the current exact owner;
- only `completed` publishes semantic progress;
- every terminal kind can still retire its own active-tool marker.

This reuses an already-established repository correctness pattern:

```
model request:
event identity + generation -> trusted lifecycle accounting

tool execution:
event identity + generation -> trusted lifecycle accounting
                                  |
                                  +-- completed -> semantic progress
```

### Why this is preferable

It avoids four new failure classes:

1. runId/sessionId reuse accidentally refreshing a successor;
2. plugin/public payload fields forging authority;
3. delayed terminal events refreshing a replacement generation;
4. worker-only coverage that misses the native embedded tool path from #166770.

This is now the strongest candidate implementation direction found in the source review.
