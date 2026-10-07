# OpenClaw #166371 — real Claude CLI proof protocol

Goal: satisfy the current ClawSweeper request with branch-specific, after-fix evidence. This must exercise the reviewed head, not the overlapping #153032 implementation.

## Required runtime shape

Use a fresh isolated state/config and the real Claude CLI runtime path with the candidate head:

- runtime: `claude-cli`
- bundle MCP mode: `claude-config-file`
- OpenClaw MCP server name: `openclaw`
- fresh session / no reused transcript
- a channel/source path where a reply must be delivered through the OpenClaw message tool

The evidence must show all three stages:

1. prompt/tool discovery exposes `mcp__openclaw__message` to Claude;
2. Claude actually calls `mcp__openclaw__message`, with no failed bare `message` call first;
3. the source/channel receives the delivered reply.

## Minimal proof task

Use a deterministic prompt that requires one visible source reply and no unrelated tools, for example:

```text
Reply to the current source with exactly:
OPENCLAW_PREFIX_PROOF_OK
Use the messaging tool to send it.
```

Do not use a prompt that can be satisfied only in stdout; the proof must exercise source delivery.

## Capture

Record redacted terminal output or a short screen recording containing:

```text
candidate head: 230f40765c1b4f48bdf26256fc3013255b760c65
runtime: claude-cli
bundle MCP mode: claude-config-file

visible tool / prompt evidence:
mcp__openclaw__message

tool call:
mcp__openclaw__message(...)

bare-tool failure:
none

delivery result:
source received OPENCLAW_PREFIX_PROOF_OK
```

If OpenClaw emits structured tool-call / delivery logs, keep the relevant lines. Remove:

- tokens / API keys;
- private channel IDs;
- usernames/phone numbers;
- private endpoints;
- unrelated conversation content.

## Negative control

On current main or a baseline build, capture only the naming mismatch needed to establish the bug:

```text
prompt guidance: message(action=send)
registered callable: mcp__openclaw__message
```

A repeated real failed tool call is useful but not required if the existing linked report already establishes the pre-fix symptom. Do not expose user data just to strengthen the baseline.

## Acceptance criteria

The after-fix run passes only if:

- candidate build SHA is visible/recorded;
- fresh Claude CLI session is used;
- `mcp__openclaw__message` is visible to the model;
- first attempted OpenClaw messaging call uses the prefixed name;
- no `No such tool available: message` occurs;
- source delivery succeeds and is independently observable.

## PR-body evidence template

```md
### Real Claude CLI behavior proof

Candidate: `230f40765c1b4f48bdf26256fc3013255b760c65`.

Fresh isolated `claude-cli` session using the production `claude-config-file` MCP bridge:

- model-visible message tool: `mcp__openclaw__message`
- first messaging call: `mcp__openclaw__message`
- bare `message` failure: none
- source delivery: succeeded; source received `OPENCLAW_PREFIX_PROOF_OK`

Redacted terminal/log capture: <artifact/link>

This exercises the candidate's prompt preparation + Claude Code MCP naming + real source-delivery path. No credentials, private identifiers, or unrelated conversation content are included.
```

## Conflict handling

Do not resolve the current merge conflict by adding a second naming owner. Preserve these invariants from the reviewed patch:

- capability/policy checks continue to use canonical bare tool identities;
- only model-visible names/guidance receive the Claude MCP prefix;
- embedded runtime behavior remains unchanged;
- other CLI backends remain unchanged unless separately proven;
- the existing `openclaw` MCP server name remains the single prefix source.

Security-sensitive approval is a maintainer action and should be requested only after the rebased head is final, because a later push invalidates the approval.

Status: source/review-derived proof protocol; no runtime execution claimed.
