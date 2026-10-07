# OpenClaw #154728 — Gateway behavior proof plan

Goal: satisfy the remaining ClawSweeper blocker with a production-boundary observation, not another unit-only merge test.

## What must be demonstrated

A run produces a proven completion first:

```json
{"status":"ok","startedAt":100,"endedAt":200}
```

A later wait-layer observation for the same run reports:

```json
{"status":"timeout","timeoutPhase":"queue","endedAt":300}
```

or:

```json
{"status":"timeout","timeoutPhase":"gateway_draining","endedAt":300}
```

After the candidate fix, the externally visible Gateway `agent.wait` / job-status projection must still report the run as completed.

## Best existing integration surface

Current main already has the right production-boundary harness in:

- `src/gateway/agent-turn/agent-wait-dedupe.test.ts`
- `src/gateway/agent-turn/agent-job.ts`

The test helper calls the real `agentHandlers["agent.wait"]` handler and the production `setGatewayDedupeEntry` / `waitForAgentJob` owners. This is stronger evidence than adding another direct `mergeAgentRunTerminalOutcome` case.

## Proposed regression

Add a case adjacent to the existing cross-source merge coverage:

1. Publish a completed lifecycle/dedupe snapshot at `endedAt: 200`.
2. Publish a later soft timeout snapshot at `endedAt: 300`, first with `timeoutPhase: "queue"`, then `"gateway_draining"`.
3. Invoke the real `agent.wait` handler through the existing `waitThroughGateway(...)` helper.
4. Assert the response remains:

```ts
expect(waiter.respond).toHaveBeenCalledWith(
  true,
  expect.objectContaining({
    runId,
    status: "ok",
    endedAt: 200,
  }),
);
```

5. Also call `waitForAgentJob({ runId, timeoutMs: 0 })` and assert the canonical stored observation remains `status: "ok"`.
6. Run the same sequence in both observation orders to prove ordering independence.

## Why this satisfies the blocker

The current review asks specifically for after-fix Gateway wait/status evidence. This path exercises:

```
terminal observations
  -> setGatewayDedupeEntry / lifecycle recording
  -> agent-job canonical snapshot merge
  -> agent.wait Gateway handler
  -> projected external status
```

That crosses the production integration boundary missing from the current PR evidence.

## Suggested command and timing capture

Run only the production-boundary file with one worker and record wall time:

```bash
/usr/bin/time -p pnpm test src/gateway/agent-turn/agent-wait-dedupe.test.ts --maxWorkers=1
```

Then run the changed owner test similarly:

```bash
/usr/bin/time -p pnpm test src/agents/agent-run-terminal-outcome.test.ts --maxWorkers=1
```

Record:
- test count;
- wall time;
- exact candidate SHA;
- one redacted `agent.wait` result for queue;
- one redacted `agent.wait` result for gateway_draining.

## Evidence wording for the PR body

> Production-boundary regression: the real Gateway `agent.wait` handler retained a previously proven completion after later queue and gateway-draining timeout observations. The canonical `waitForAgentJob` snapshot also remained completed in both observation orders. The test uses the existing Gateway handler and agent-job state owners; only provider execution is outside this proof. Single-worker wall time: <fill measured value>.

## Scope discipline

Do not change:
- hard timeout precedence;
- cancellation precedence;
- unattributed delivery timeout semantics;
- `agent.wait` interruption behavior itself.

The proof should validate the existing candidate, not broaden the patch.
