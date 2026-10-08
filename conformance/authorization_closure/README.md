# Authorization Closure Conformance Harness

Status: experimental v0.1

This harness evaluates whether an authorization/revocation flow has actually reached a consequential execution boundary.

It is intentionally framework-neutral.

A runtime does not need to adopt a specific SDK. It only needs to export a normalized trace of authorization state, revocation state, consequential sinks, and committed effects.

The harness classifies the observed closure state as:

- CLOSED — every declared consequential path is proven closed and no prohibited effect committed after revocation became effective at its sink.
- PARTIAL — at least one path is proven closed, but at least one declared path remains open.
- UNKNOWN — the evidence is insufficient to prove closure or an open path.
- VIOLATION — a consequential effect committed after revocation became effective at that sink.

## Core distinction

historically valid authorization artifact
!=
current execution authority

and:

revocation recorded at authority
!=
revocation effective at sink
!=
closure

The harness turns those distinctions into executable evidence.

## Trace model

Each JSON trace contains:

- operation_id
- revocation_id
- sinks
- events

Supported event types in v0.1:

- authorization_accepted
- derived_credential_issued
- work_queued
- revocation_recorded
- revocation_effective
- sink_closed
- effect_committed
- effect_absent
- evidence_unavailable

Important semantics:

- revocation_recorded means the authority accepted or recorded the revocation.
- revocation_effective is sink-specific and means the contract says new consequential effects at that sink are no longer authorized.
- sink_closed means there is positive closure evidence for that path.

Do not collapse these facts.

## Propagation windows

An effect after revocation_recorded but before revocation_effective at a sink is not automatically a VIOLATION. It may fall inside a documented propagation window.

An effect after revocation_effective at that sink is VIOLATION.

UNKNOWN is preferred over inventing success when evidence is missing.

## Run

```bash
python conformance/authorization_closure/runner.py   conformance/authorization_closure/fixtures/closed.json
```

Run tests:

```bash
python -m unittest conformance.authorization_closure.test_runner
```

## Adapter strategy

Phase 1 uses exported traces only.

Future adapters can map native events from:
- OpenClaw admission / credential events;
- Open Agent Auth resource-server validation;
- OpenA2A AAP broker resolution;
- Temporal Workflow / Activity history;
- MCP gateways;
- payment or cloud-control systems.

The conformance contract should remain stable even if adapter APIs differ.

## Contribution rule

Add a scenario only when it corresponds to:
1. a real implementation behavior or failure;
2. a deterministic evidence path;
3. a consequential sink;
4. a portable authority/closure distinction.

Do not grow this into an abstract taxonomy.

## Current evidence status

- portable invariant: established across multiple runtimes;
- reference evaluator: available;
- schema: available in this directory;
- CI self-test: available;
- real third-party adapter: not yet;
- independent reuse: not yet;
- standards adoption: not yet.


## Independent-runtime evidence adapters

The harness now consumes evidence from two independent systems:

### OpenClaw

Adapter:
- `adapters/openclaw_exec_launch_policy.py`

Evidence source:
- merged OpenClaw launch-policy regression around native child / PTY / Claude process construction authority.

Current mapped result for the upstream pre-launch revocation case:
- `CLOSED`

### Open Agent Auth

Adapter:
- `adapters/open_agent_auth_revocation.py`

Evidence sources:
- upstream `InMemoryTokenRevocationServiceTest` proves authority-side `revoke -> isRevoked`;
- current AOAT / Resource Server source review does not establish that the Resource Server consumes that revocation state.

Current mapped result:
- `UNKNOWN`

This is intentional. Missing sink-enforcement evidence must not be promoted into either `CLOSED` or `VIOLATION`.

## CI

Workflow:
- `.github/workflows/authorization-closure-conformance.yml`

The workflow runs the core evaluator tests plus both independent-runtime adapter tests whenever this conformance directory changes.

A configured CI workflow is not the same as third-party adoption. The stronger milestones remain native telemetry, an external adapter contribution, or standards/conformance reuse.
