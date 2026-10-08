# OpenClaw #166770 — candidate implementation boundary

Status: source-reviewed design artifact, not compiled or executed.

This file narrows the remaining implementation choice after the broader analysis in:
`openclaw-166770-semantic-progress-from-settled-tool-execution.md`.

## Proven facts

### 1. Terminal truth already exists

`agent-tools.before-tool-call.wrapper.ts` classifies real tool outcomes as:

- `tool.execution.completed`
- `tool.execution.error`
- `tool.execution.blocked`

after the implementation returns and after `resolveToolResultTerminalDiagnostic()`.

So #166770 does **not** need to infer success from model `stopReason`.

### 2. Exact embedded-run ownership already exists

The embedded runner owns a `DiagnosticEmbeddedRunOwner`, including an opaque generation object.

That is the correct authority identity. Matching `runId` / `sessionId` is not equivalent.

### 3. OpenClaw already has the provenance mechanism

`diagnostic-model-request-provenance.ts` binds exact event object identity to an owner generation using a private `WeakMap<object, provenance>`.

This establishes a repository-native pattern:

```
core runtime creates event
+ privately binds generation
+ diagnostic dispatcher consumes object-identity provenance
+ listeners receive trusted metadata
```

### 4. The watchdog already has the right behavioral test seam

`src/logging/diagnostic.test.ts` contains the repeated-request semantic-progress regression and exact-owner/stale-owner fixtures needed to prove the fix.

## Candidate code family

### A. Add core-private tool lifecycle provenance

Mirror the model-request pattern in a small module such as:

`src/infra/diagnostic-tool-execution-provenance.ts`

Conceptually:

```ts
export type CoreToolExecutionLifecycleProvenance = Readonly<{
  generation: CoreModelRequestOwnerGeneration;
  phase: "started" | "completed" | "error" | "blocked";
}>;

const events = new WeakMap<object, CoreToolExecutionLifecycleProvenance>();

export function markCoreToolExecutionLifecycleDiagnosticEvent<T extends ToolLifecycleEvent>(
  event: T,
  provenance: CoreToolExecutionLifecycleProvenance,
): T {
  events.set(event, provenance);
  return event;
}

export function consumeCoreToolExecutionLifecycleDiagnosticEvent(
  event: object,
): CoreToolExecutionLifecycleProvenance | undefined {
  const provenance = events.get(event);
  events.delete(event);
  return provenance;
}
```

Exact names can follow local conventions.

### B. Thread provenance through trusted diagnostic metadata

Extend the existing trusted diagnostic metadata plumbing in
`src/infra/diagnostic-events.ts` exactly as model-request provenance is threaded.

Required property:

> payload fields cannot create or forge the owner generation.

### C. Bind the embedded run's existing owner generation to ordinary tool execution

This is the **only remaining implementation-choice seam**.

The canonical embedded path already owns `DiagnosticEmbeddedRunOwner`, but
`HookContext` is plugin-facing enough that placing an authority-bearing generation directly on it
would be an undesirable API expansion.

The implementation should therefore use a host-private path. Candidate families include:

1. a private execution-wrapper/metadata carrier attached to the wrapped tool;
2. a host-owned closure created while preparing the run's tool surface;
3. an AsyncLocalStorage-style private owner carrier around admitted tool execution.

Whichever is smallest in current main, it must satisfy:

```
plugin/public hook context cannot forge or retain owner authority
+
terminal event receives the exact generation of the run that admitted execution
```

Do not solve this by adding `diagnosticOwner` to ordinary plugin-visible HookContext.

### D. Run-activity accounting

When trusted tool terminal metadata reaches run activity:

```
if provenance.phase === "completed"
and generation is the exact current embedded-run owner
then:
    retire tool marker
    publish semantic progress
else:
    retire the appropriate marker only
```

`error` and `blocked` remain terminal work, but do not prove successful semantic progress.

A delayed completion from a replaced/closed generation must not reset the successor's watchdog clock.

## Regression matrix

### Positive

- current exact owner
- successful tool execution
- later enclosing model result ends in `length`
- repeated-request watchdog must not recover for no semantic progress

### Negative

- error terminal
- blocked terminal
- start without terminal
- completion from closed owner generation
- completion from replaced owner generation
- unbound delayed public terminal with matching run/session payload only

## Explicitly rejected patch shapes

- `stopReason === "length"` + normalized call => progress
- any tool terminal => progress
- matching runId/sessionId => current authority
- worker-only `markDiagnosticOwnedToolActivity()` change
- plugin-visible diagnostic owner generation

## Remaining unknown

The exact host-private injection point that binds `DiagnosticEmbeddedRunOwner.generation` to the
ordinary wrapped tool execution has not been proven from the inspected source yet.

Everything else in the repair is now bounded.

That means the next source pass should answer one question only:

> Where is the narrowest host-private point that already has both the prepared tool and the embedded run's diagnostic owner?

Once that seam is identified, the implementation can remain a small provenance extension rather
than a new lifecycle framework.
