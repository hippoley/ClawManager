# OpenClaw #161046 — real preflight-to-same-turn proof without external provider credentials

Target head: `303ba6bedc8622068c8ce6c5c497213623f3e617`

## Review blocker

ClawSweeper found no concrete patch defect. The remaining blocker is one production-boundary fact:

```text
message arrives while host preflight compaction is active
→ preflight succeeds
→ embedded backend becomes ready
→ that exact message is injected into the same active turn
```

The previous isolated Gateway attempt did not establish this because host compaction failed before
backend readiness due to missing provider API authentication.

## Existing repository seam that avoids external credentials

Current OpenClaw already has a stronger local-provider boundary in:

`src/auto-reply/reply/agent-runner-compaction-fallback.test.ts`

That test uses:

- a real HTTP server on `127.0.0.1`;
- an OpenAI-compatible provider configured through normal model config;
- real preflight compaction;
- the native compaction delegate;
- the real foreground embedded run.

The provider is local, but transport/config/provider dispatch are production paths. No external API
key is required.

The real embedded message-injection owner is also production code:

`src/agents/embedded-agent-runner/run/attempt-stream-prepare.ts`

It publishes `messageInjectionV2.queueMessage`, whose implementation calls:

```ts
steerActiveSessionWithOptionalDeliveryWait(...)
  -> activeSession.steer(...)
```

Therefore the missing proof can combine a real Gateway client with the existing local HTTP provider
rather than using the synthetic backend from
`agent-runner-preflight-steering.integration.test.ts`.

## Required proof shape

Use a real in-process Gateway test harness with a real authenticated `GatewayClient`, and configure
the target agent/model to the local OpenAI-compatible HTTP endpoint.

### 1. Seed a session that requires host preflight compaction

Reuse the existing compaction-fallback fixture shape:

- real session state;
- enough prior messages/tokens to force preflight compaction;
- provider `test-provider/test-model`;
- normal OpenAI-compatible base URL pointing to the local HTTP server.

### 2. Hold a successful summary request at the provider

The local provider server should distinguish summarization traffic the same way the existing test
does (request body contains the summarization system prompt).

Instead of returning 408, use a deferred successful response:

```text
summary HTTP request arrives
→ signal preflightEntered
→ hold response open
```

At this point the active reply operation must still be the same operation, in preflight, without an
attached foreground backend.

### 3. Send the second user message through the real Gateway transport

Do **not** call `runReplyAgent()` directly for the second message.

Use the real Gateway client / chat or agent RPC used by the product surface, on the same session and
queue key.

Record:

```text
second message submitted after preflightEntered
first operation has not completed
foreground provider request has not started yet
```

The second message should remain parked as a pending steer for that exact operation.

### 4. Let preflight succeed

Return a valid summary response from the local OpenAI-compatible provider.

Then observe:

```text
host compaction commits
→ operation leaves preflight
→ foreground embedded run starts
→ attempt-stream-prepare attaches the real embedded backend
→ waiting steer re-resolves against the same operation
→ messageInjectionV2.queueMessage
→ activeSession.steer(secondMessage)
```

## What counts as same-turn proof

Do not use only internal registry state.

Capture at least two independent observations:

1. **Injection ownership**
   - same reply operation/run id before and after preflight;
   - second message receives an accepted steering disposition from the real Gateway path;
   - no successor reply operation is created for that message before the active turn settles.

2. **Embedded-session consumption**
   - a transcript / diagnostic event from the active embedded session records the second user input
     under the same run/turn;
   - or a provider/runtime-visible continuation proves the active session consumed the steer before
     terminal completion.

The strongest compact trace is:

```text
run=<R> phase=preflight_compacting
summary request entered
message=<M2> submitted via Gateway while run=<R> still active
summary success
run=<R> backend attached
message=<M2> steering accepted by run=<R>
message=<M2> committed/consumed before run=<R> terminal
run=<R> terminal
no successor run owns <M2>
```

## Local HTTP provider behavior

The local provider can make the active-turn observation deterministic.

A practical sequence:

```text
summary request:
  hold until M2 has been submitted
  then return a valid summary

foreground request:
  begin SSE response
  keep the turn open long enough for steering delivery
  record any follow-on provider/runtime request or embedded-session transcript resulting from M2
  finish only after same-turn steering has been observed
```

Do not assert that the second message must appear in the *initial* foreground HTTP request body; it
arrives after backend attachment and is delivered through the embedded session steering owner.

## Red/green control

Run the same proof on:

1. candidate `303ba6bedc8622068c8ce6c5c497213623f3e617`;
2. current-main behavior / candidate with the preflight wait owner reverted.

Expected:

```text
candidate:
M2 submitted during successful preflight
→ accepted into active run R
→ consumed before R terminal

pre-fix:
M2 submitted during preflight
→ no injection target
→ falls back / waits until R ends
→ later turn owns M2
```

## Why this is stronger than the current evidence

The existing PR integration test is valuable but explicitly says:

```text
provider/backend transport remains synthetic
```

This proposed proof uses:

- real Gateway transport;
- real session state;
- real host preflight;
- real OpenAI-compatible HTTP provider transport;
- real embedded backend registration;
- real `messageInjectionV2.queueMessage`;
- real `activeSession.steer`.

The only controlled part is the local provider's deterministic response timing.

## Acceptance artifact

A redacted log is sufficient:

```text
Head: 303ba6bedc8622068c8ce6c5c497213623f3e617
Provider transport: local OpenAI-compatible HTTP endpoint
Gateway transport: real authenticated Gateway client

T+0      run R enters preflight_compacting
T+...    summary HTTP request received
T+...    second message M2 sent through Gateway
T+...    summary HTTP 200 returned
T+...    run R backend attached
T+...    M2 steer accepted by run R
T+...    M2 transcript/runtime consumption observed in run R
T+...    run R completes
successor run for M2: none
```

No key, local port, filesystem path, private message text, or personal data needs to be published.

## Scope

This proves the exact review gap. It does not claim:

- universal queue FIFO;
- paid-provider compatibility;
- compaction latency improvement;
- steering after owner replacement/cancellation.

Those remain covered by existing focused tests.

Status: source-reviewed against #161046 plus current repository local-provider and embedded-injection
owners; not executed by `hippoley`.
