# OpenClaw #166511 — preserve parent continuation hold across child traffic and host suspend

Target head: `0a7719fe7ed5c491cd93d0feab48ac832bd45c9b`

## Blocker 1: forwarded child output must not release the parent hold

Current `execute-plugin.ts` keeps the continuation hold only for selected `system` bookkeeping.
That means child records such as:

- `assistant` with `parent_tool_use_id`;
- `user` tool-result records with `parent_tool_use_id`;
- `stream_event` records with `parent_tool_use_id`;

fall through as if the parent conversation resumed.

The repository already has the canonical classifier:

```ts
isClaudeSubagentRecord(parsed)
```

from `src/agents/cli-output-records.ts`.

### Minimal repair

Import and reuse it in `execute-plugin.ts`.

Conceptually:

```ts
const childTraffic = isClaudeSubagentRecord(next.value);

const awaitingContinuation =
  next.value.type === "result"
    ? next.value.openclaw_interim_result === true
    : outstanding.awaitingContinuation &&
      (childTraffic ||
        (next.value.type === "system" &&
          (next.value.subtype !== "task_notification" ||
            (typeof next.value.task_id === "string" &&
              subagentTaskIds.has(next.value.task_id)))));
```

This preserves the existing special case for subagent-owned task notifications and extends the same
parent/child distinction to forwarded assistant/user/stream-event records.

Do not classify arbitrary non-system records as child traffic without `parent_tool_use_id`; those
are parent-lane activity and should release the hold.

## Blocker 2: diagnostics must follow the watchdog's suspend-adjusted overall deadline

The watchdog already owns the correct active-time budget.

On a long host suspend it computes:

```text
elapsed wall gap
- suspend credit
= active elapsed

overallActiveRemainingMs -= active elapsed
```

and therefore extends the absolute effective deadline by the suspend duration.

The problem is only publication: `onContinuationHoldChange()` receives
`watchdog.overallDeadlineAtMs()` once when the hold starts.

### Smallest owner-correct repair

Let the watchdog publish only actual overall-deadline changes caused by suspend accounting.

In `execute-plugin-watchdog.ts`:

```ts
onOverallDeadlineChange?: (deadlineAtMs: number | undefined) => void;
```

Factor the existing projection:

```ts
const overallDeadlineAtMs = () =>
  overallActiveRemainingMs === undefined
    ? undefined
    : lastTickAtMs + overallActiveRemainingMs;
```

After a tick applies a non-zero suspend credit and updates
`overallActiveRemainingMs`, publish:

```ts
if (suspendedMs > 0) {
  params.onOverallDeadlineChange?.(overallDeadlineAtMs());
}
```

Return the same helper from the watchdog API.

In `execute-plugin.ts`, wire:

```ts
onOverallDeadlineChange: (deadlineAtMs) => {
  if (outstanding.awaitingContinuation) {
    params.onContinuationHoldChange?.(deadlineAtMs);
  }
},
```

This keeps diagnostics synchronized only while the parent continuation hold is active.

No second deadline calculation is introduced. The watchdog remains the single owner of
suspend-adjusted execution time.

## Why not refresh diagnostics on every watchdog tick

A per-second `holdUntil()` write is unnecessary. The absolute overall deadline is stable during
normal active execution:

```text
lastTickAtMs += activeElapsed
overallRemaining -= activeElapsed
sum remains constant
```

It changes only when active time differs from wall time, principally host suspension.

Therefore publishing on suspend credit is sufficient and avoids noisy diagnostic mutations.

## Regression A — forwarded subagent output

Extend `execute.continuation-hold.test.ts`.

After the interim answer, emit real child-shaped records before the background wait:

```ts
yield {
  type: "assistant",
  parent_tool_use_id: "agent-call",
  message: { role: "assistant", content: [{ type: "text", text: "child progress" }] },
};

yield {
  type: "stream_event",
  parent_tool_use_id: "agent-call",
  event: {
    type: "content_block_start",
    content_block: { type: "tool_use", id: "child-tool", name: "Bash", input: {} },
  },
};
```

Then advance beyond both ordinary no-output and diagnostics grace windows.

Required:

```text
recoverStuckSession not called
run not aborted
parent continuation hold still active
```

Finally emit a genuine parent continuation event and verify ordinary watchdog behavior resumes.

The regression should fail on the current head because the first child assistant record clears
`awaitingContinuation`.

## Regression B — host suspension

Use a controllable watchdog clock plus the existing diagnostic fake clock.

Shape:

```text
interim answer
→ continuation hold starts with deadline D0

simulate host suspend > HOST_SUSPEND_TICK_THRESHOLD_MS
→ watchdog credits frozen interval
→ effective overall deadline becomes D1 > D0
→ diagnostics hold must also become D1

advance wall clock to just after D0 but before D1
→ recoverStuckSession must NOT fire

advance active time to D1
→ overall timeout may settle the held answer normally
```

The key assertion is on the diagnostic snapshot's
`activeBackendLivenessDeadlineAtMs`, not only on the watchdog.

## Production patch skeleton

```diff
diff --git a/src/agents/cli-runner/execute-plugin.ts b/src/agents/cli-runner/execute-plugin.ts
@@
+import { isClaudeSubagentRecord } from "../cli-output-records.js";
@@
       const awaitingContinuation =
         next.value.type === "result"
           ? next.value.openclaw_interim_result === true
           : outstanding.awaitingContinuation &&
-            next.value.type === "system" &&
-            (next.value.subtype !== "task_notification" ||
-              (typeof next.value.task_id === "string" && subagentTaskIds.has(next.value.task_id)));
+            (isClaudeSubagentRecord(next.value) ||
+              (next.value.type === "system" &&
+                (next.value.subtype !== "task_notification" ||
+                  (typeof next.value.task_id === "string" &&
+                    subagentTaskIds.has(next.value.task_id)))));
@@
       onOverallTimeout: () => {
         ...
       },
+      onOverallDeadlineChange: (deadlineAtMs) => {
+        if (outstanding.awaitingContinuation) {
+          params.onContinuationHoldChange?.(deadlineAtMs);
+        }
+      },
diff --git a/src/agents/cli-runner/execute-plugin-watchdog.ts b/src/agents/cli-runner/execute-plugin-watchdog.ts
@@
+    onOverallDeadlineChange?: (deadlineAtMs: number | undefined) => void;
@@
+  const overallDeadlineAtMs = () =>
+    overallActiveRemainingMs === undefined
+      ? undefined
+      : lastTickAtMs + overallActiveRemainingMs;
@@
     if (overallActiveRemainingMs !== undefined) {
       overallActiveRemainingMs -= activeElapsedMs;
       ...
     }
+    if (suspendedMs > 0) {
+      params.onOverallDeadlineChange?.(overallDeadlineAtMs());
+    }
@@
-    overallDeadlineAtMs: () =>
-      overallActiveRemainingMs === undefined ? undefined : lastTickAtMs + overallActiveRemainingMs,
+    overallDeadlineAtMs,
```

Exact placement should preserve the existing early `onOverallTimeout()` return: do not publish a
future hold after the overall budget has already expired.

## Acceptance

Use the reviewer's exact owning suites:

```bash
pnpm test   src/agents/cli-runner/execute.continuation-hold.test.ts   src/agents/cli-runner/execute.background-work.test.ts   src/agents/cli-runner/execute-plugin.test.ts   src/logging/diagnostic-backend-liveness.test.ts   --maxWorkers=1

pnpm test   src/agents/cli-output-jsonl.test.ts   src/agents/cli-runner/execute.compaction-watchdog.test.ts   src/agents/cli-runner/execute.pending-cancellation.test.ts   --maxWorkers=1

pnpm check:changed
git diff --check
```

Status: source-reviewed against `0a7719fe7ed5c491cd93d0feab48ac832bd45c9b`; not executed by
`hippoley`.
