# OpenClaw #166399 — real Gateway policy-refusal proof contract

Target head: `5030d360761eb270d254e821d2e876be917f6b09`

Purpose: satisfy the current ClawSweeper proof gate with after-fix evidence from a real Gateway/runtime setup, not the existing fixture-only composition.

## Behavior under proof

A deliberate `before_agent_run` policy refusal must:

1. return a safe replacement reply;
2. retain the existing Ask OpenClaw conversation;
3. preserve the same internal session identity for the next turn;
4. allow a subsequent permitted turn to run normally;
5. keep the blocked private input and internal policy reason out of projected history/transcript;
6. retain only the safe replacement representation needed for continuity.

## Minimal real scenario

Use a fresh isolated OpenClaw state and a real Gateway.

Install/configure a test policy hook whose behavior is:

```text
if prompt == POLICY_BLOCK_SECRET_42:
    outcome = block
    internal reason = INTERNAL_POLICY_REASON_42
    safe message = Please send an allowed request.
else:
    pass
```

Use a real configured system-agent inference route. Do not stub the embedded runner/model or transcript store.

### Turn 1 — blocked

Send through the real Ask OpenClaw Gateway surface:

```text
POLICY_BLOCK_SECRET_42
```

Required external observation:

```text
reply contains:
Please send an allowed request.
```

The request must complete as a normal safe reply, not as an inference-unavailable transport error.

### Turn 2 — allowed

On the same Ask OpenClaw conversation, send:

```text
Allowed follow-up: reply exactly OPENCLAW_POLICY_CONTINUITY_OK
```

Required observation:

```text
OPENCLAW_POLICY_CONTINUITY_OK
```

## Identity evidence

Capture the system-agent session identity from redacted Gateway/runtime logs before/after both turns.

Required invariant:

```text
turn 1 sessionId == turn 2 sessionId
turn 1 sessionKey == turn 2 sessionKey
```

If the logs expose the backing conversation/session-manager identity, record that too. Do not publish secrets, account IDs, phone numbers, or private channel identifiers.

The key point is to prove the Gateway did not dispose and recreate the conversation after `hook_block`.

## History safety evidence

After turn 2, inspect the projected Ask OpenClaw history/transcript using the production history/debug surface available in the test environment.

Required assertions:

```text
contains:
Please send an allowed request.

does not contain:
POLICY_BLOCK_SECRET_42
INTERNAL_POLICY_REASON_42
```

The allowed second user turn and normal assistant response may appear as usual.

This proves the patch's replacement-only history contract, not merely conversation survival.

## Failure-mode control

Also capture that a genuine inference/runtime failure still follows the existing disposal/failure behavior. This can be a separate bounded control if the environment already has a safe way to force an inference error.

Do not weaken the proof by converting all terminal failures into retained conversations; only typed policy refusals are in scope.

## Evidence bundle

A sufficient redacted bundle should include:

```text
candidate head: 5030d360761eb270d254e821d2e876be917f6b09
Gateway: real process
system-agent runtime/model: real configured route

TURN 1
input: <redacted blocked marker>
result: safe policy reply
transport/inference-unavailable error: none
sessionId: <redacted stable token>

TURN 2
input: allowed follow-up
result: OPENCLAW_POLICY_CONTINUITY_OK
sessionId: same stable token

HISTORY
safe replacement present: yes
blocked private input present: no
internal policy reason present: no
```

Terminal logs are sufficient if they establish all invariants. A short screen recording plus redacted logs is stronger.

## PR-body template

```md
### Real Gateway policy-refusal continuity proof

Candidate: `5030d360761eb270d254e821d2e876be917f6b09`.

Ran a fresh isolated real Gateway with a real configured system-agent inference route and a deterministic `before_agent_run` blocking hook.

- blocked turn returned the safe replacement as a normal successful Ask OpenClaw reply;
- conversation was retained;
- follow-up allowed turn completed normally;
- the same internal system-agent session identity was observed across both turns;
- projected history contains the safe replacement;
- projected history does not contain the blocked private input or internal policy reason.

Allowed follow-up marker: `OPENCLAW_POLICY_CONTINUITY_OK`.

Redacted logs / recording: <link>

This is after-fix runtime evidence for the reviewed branch, not the fixture-only regression.
```

## Scope

This proof must not claim or imply any persistent-schema migration. The reviewed patch changes refusal propagation and history projection semantics only.

Status: proof contract derived from the reviewed production path and regression; no runtime execution is claimed here.
