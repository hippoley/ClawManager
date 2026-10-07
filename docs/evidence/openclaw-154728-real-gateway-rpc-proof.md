# OpenClaw #154728 — real Gateway RPC proof after the handler-level regression

Target head: `1edb04386eb9b5cf9d55410dc87a32d095890d06`

## Why another proof is still needed

Revision 5 accepts the correctness of the focused patch and the new
`agent-wait-dedupe.test.ts` regression, but still marks proof 2/6 because the test invokes the
registered `agent.wait` handler with a synthetic request context/response spy.

The remaining gate is specifically:

```text
real Gateway transport client
→ agent.wait RPC
→ production canonical run state
→ ok/200 after later queue or gateway_draining timeout observation
```

The repository already has all required infrastructure.

## Reuse existing real transport harness

Use:

- `startGatewayWithClient()` from `src/gateway/test-helpers.e2e.ts`
- the returned real `GatewayClient`
- production `setGatewayDedupeEntry()`
- real `client.request("agent.wait", ...)`

This avoids creating a new protocol client or depending on a model/provider.

## Minimal live-RPC scenario

A focused e2e/live test can live next to the existing Gateway wait regression or as a one-shot
proof script.

Pseudo-exact structure:

```ts
import { randomUUID } from "node:crypto";
import { expect, it } from "vitest";
import type { OpenClawConfig } from "../config/types.openclaw.js";
import {
  setGatewayDedupeEntry,
  waitForAgentJob,
} from "./agent-turn/agent-job.js";
import {
  disconnectGatewayClient,
  startGatewayWithClient,
} from "./test-helpers.e2e.js";

it.each(["queue", "gateway_draining"] as const)(
  "preserves completed run over later %s timeout through real Gateway RPC",
  async (timeoutPhase) => {
    const token = `terminal-proof-${randomUUID()}`;
    const cfg = {
      gateway: { auth: { mode: "token", token } },
      plugins: { slots: { memory: "none" } },
      tools: { deny: ["*"] },
    } satisfies OpenClawConfig;

    const gateway = await startGatewayWithClient({
      cfg,
      token,
      scopes: ["operator.admin", "operator.read", "operator.write"],
    });

    try {
      await gateway.server.startupSettled;

      const runId = randomUUID();
      const dedupe = new Map();

      setGatewayDedupeEntry({
        dedupe,
        key: `agent:${runId}`,
        entry: {
          ts: 200,
          response: {
            payload: {
              status: "ok",
              startedAt: 100,
              endedAt: 200,
            },
          },
        },
      });

      setGatewayDedupeEntry({
        dedupe,
        key: `agent:${runId}`,
        entry: {
          ts: 300,
          response: {
            payload: {
              status: "timeout",
              startedAt: 100,
              endedAt: 300,
              timeoutPhase,
            },
          },
        },
      });

      const rpc = await gateway.client.request<{
        status?: string;
        endedAt?: number;
      }>(
        "agent.wait",
        { runId, timeoutMs: 5_000 },
        { timeoutMs: 10_000 },
      );

      const internal = await waitForAgentJob({
        runId,
        timeoutMs: 0,
        source: "agent",
      });

      expect(rpc).toMatchObject({ status: "ok", endedAt: 200 });
      expect(internal).toMatchObject({ status: "ok", endedAt: 200 });

      process.stdout.write(
        JSON.stringify({
          timeoutPhase,
          rpc,
          internal: internal && {
            status: internal.status,
            endedAt: internal.endedAt,
          },
        }) + "\n",
      );
    } finally {
      await disconnectGatewayClient(gateway.client);
      await gateway.server.close({ reason: "terminal-outcome proof cleanup" });
    }
  },
);
```

## Important implementation note: use the Gateway's actual dedupe owner

If `startGatewayWithClient()` does not expose the exact dedupe map used by the running server,
do **not** create an unrelated local Map as in the schematic above.

Instead use the same test seam already employed by the in-process Gateway fixtures to reach the
server-owned dedupe map, or publish the observations through the production owner that writes into
that map.

The acceptance criterion is about ownership, not syntax:

```text
setGatewayDedupeEntry
and
agent.wait RPC
must observe the same canonical agent-job state
```

A local throwaway Map plus an unrelated server would not qualify.

The current handler regression already proves the merger; this live proof must prove the transport
projection.

## Better fixture shape if the server-owned dedupe map is available

The strongest form is:

```text
startGatewayWithClient()
→ server-owned request context / dedupe map
→ setGatewayDedupeEntry(completed@200)
→ setGatewayDedupeEntry(soft-timeout@300)
→ GatewayClient.request("agent.wait")
→ {status:"ok", endedAt:200}
```

Run for both:

- `queue`
- `gateway_draining`

Also run the same harness with only the merge-owner file reverted to current main and record:

```text
{status:"timeout", endedAt:300}
```

That provides the reviewer-requested red/green real transport evidence.

## Output to paste into PR body

A sufficient redacted artifact is:

```text
Head: 1edb04386eb9b5cf9d55410dc87a32d095890d06
Transport: real Gateway WebSocket client

queue:
  candidate RPC: {"status":"ok","endedAt":200}
  reverted merge-owner RPC: {"status":"timeout","endedAt":300}

gateway_draining:
  candidate RPC: {"status":"ok","endedAt":200}
  reverted merge-owner RPC: {"status":"timeout","endedAt":300}
```

No IP address, token, filesystem path, or private endpoint needs to be published.

## Dependency/security guard interpretation

The PR now has dependency-graph and security-sensitive approval guards after merging current
`main`, but ClawSweeper Revision 5 independently isolates the focused fix to:

- `src/agents/agent-run-terminal-outcome-merge.ts`
- `src/agents/agent-run-terminal-outcome.test.ts`
- `src/gateway/agent-turn/agent-wait-dedupe.test.ts`

The broad package-manifest and secret/auth files are incorporated upstream/main differences, not
new fix-specific edits.

Do not claim the approvals are unnecessary: the platform guard still requires a maintainer decision
on the current head. The accurate statement is that they are **merge-composition approvals**, not
evidence that this focused terminal-outcome fix itself introduced dependency or secret-handling
logic.

Status: source-reviewed against `1edb04386eb9b5cf9d55410dc87a32d095890d06`; not executed.
