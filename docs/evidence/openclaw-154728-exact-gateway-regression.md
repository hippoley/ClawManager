# OpenClaw #154728 — exact Gateway regression patch

Target branch: `gokay-ai/openclaw:cursor/soft-timeout-completed-precedence-85bc`

Add this case to `src/gateway/agent-turn/agent-wait-dedupe.test.ts` near the existing cross-source merge coverage:

```ts
  it.each(["queue", "gateway_draining"] as const)(
    "keeps proven completion over later %s timeout through agent.wait",
    async (timeoutPhase) => {
      for (const completionFirst of [true, false]) {
        const runId = `run-soft-timeout-${timeoutPhase}-${completionFirst}`;
        const dedupe = new Map<string, DedupeEntry>();

        const completed = () =>
          setGatewayDedupeEntry({
            dedupe,
            key: `agent:${runId}`,
            entry: {
              ts: 200,
              ok: true,
              payload: {
                runId,
                status: "ok",
                startedAt: 100,
                endedAt: 200,
              },
            },
          });

        const laterSoftTimeout = () =>
          setGatewayDedupeEntry({
            dedupe,
            key: `agent:${runId}`,
            entry: {
              ts: 300,
              ok: false,
              payload: {
                runId,
                status: "timeout",
                startedAt: 100,
                endedAt: 300,
                timeoutPhase,
                providerStarted: false,
              },
            },
          });

        for (const publish of completionFirst
          ? [completed, laterSoftTimeout]
          : [laterSoftTimeout, completed]) {
          publish();
        }

        const waiter = waitThroughGateway({ runId, timeoutMs: 0 }, "agent");
        await waiter.promise;
        expect(waiter.respond).toHaveBeenCalledWith(
          true,
          expect.objectContaining({
            runId,
            status: "ok",
            endedAt: 200,
          }),
        );

        await expect(waitForAgentJob({ runId, timeoutMs: 0, source: "agent" })).resolves.toMatchObject({
          status: "ok",
          endedAt: 200,
        });
      }
    },
  );
```

Notes:
- This uses the real `agentHandlers["agent.wait"]` path through the file's existing helper.
- It exercises production `setGatewayDedupeEntry`, canonical agent-job state, and `waitForAgentJob`.
- It covers both soft timeout phases requested by review.
- It checks both observation orders.
- It deliberately does not alter hard timeout, cancellation, or delivery settlement semantics.

Suggested timing capture:

```bash
/usr/bin/time -p pnpm test src/gateway/agent-turn/agent-wait-dedupe.test.ts --maxWorkers=1
/usr/bin/time -p pnpm test src/agents/agent-run-terminal-outcome.test.ts --maxWorkers=1
```

Caveat: this is source-reviewed against the PR branch, not executed in this environment.


## Self-audit: waitForAgentJob signature

Verified against current main:

```ts
export async function waitForAgentJob(params: {
  runId: string;
  timeoutMs: number;
  source?: "agent" | "chat";
  ...
})
```

So the proposed `waitForAgentJob({ runId, timeoutMs: 0, source: "agent" })` call is
type-correct. Current repository tests also use `source: "agent"` directly in
`agent-job.execution-settlement.test.ts`.

No correction is required for this part of the regression.
