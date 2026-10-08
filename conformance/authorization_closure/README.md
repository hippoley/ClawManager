# Authorization Closure Conformance Harness

Status: experimental v0.1

This harness evaluates whether an authorization/revocation flow has actually reached a consequential execution boundary.

It is intentionally framework-neutral.

A runtime does not need to adopt a specific SDK. It only needs to export a normalized trace of:
- authorization state changes;
- derived credentials / queued work;
- consequential sinks;
- revocation observations;
- committed effects.

The harness then classifies the observed closure state as:

- CLOSED — every declared consequential path is proven closed and no prohibited effect committed after revocation became effective.
- PARTIAL — at least one path is proven closed, but at least one declared path remains open.
- UNKNOWN — the evidence is insufficient to prove closure or an open path.
- VIOLATION — a consequential effect committed after the relevant revocation was effective for that path.

## Why this exists

Across independent systems the same failure shape keeps recurring:

historically valid authorization artifact
!=
current execution authority

and:

revocation recorded at authority
!=
revocation enforced at consequential sink

The harness is designed to turn that distinction into executable evidence.

## Minimal trace format

Each JSON trace contains:

- operation_id: logical operation identity
- revocation_id: authority transition identity
- sinks: declared consequential sinks that must close
- events: ordered evidence records

Supported event types in v0.1:

- authorization_accepted
- derived_credential_issued
- work_queued
- revocation_recorded
- sink_closed
- effect_committed
- effect_absent
- evidence_unavailable

Example:

```json
{
  "operation_id": "pay-42",
  "revocation_id": "rev-9",
  "sinks": ["resource-server", "queued-worker"],
  "events": [
    {"seq": 1, "type": "authorization_accepted"},
    {"seq": 2, "type": "work_queued", "sink": "queued-worker"},
    {"seq": 3, "type": "revocation_recorded"},
    {"seq": 4, "type": "sink_closed", "sink": "resource-server"},
    {"seq": 5, "type": "sink_closed", "sink": "queued-worker"}
  ]
}
```

## Semantics

The harness deliberately distinguishes:

revocation request
!=
revocation recorded/effective
!=
closure

A trace can only be CLOSED when every declared sink has closure evidence.

A cryptographically valid or unexpired credential is not treated as proof of current authority.

If a consequential effect commits after revocation_recorded and before its sink is closed, the trace is VIOLATION.

UNKNOWN is preferred over inventing success when evidence is missing.

## Run

Python standard library only:

```bash
python conformance/authorization_closure/runner.py   conformance/authorization_closure/fixtures/closed.json
```

Run self-tests:

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

The conformance contract remains stable even if adapter APIs differ.

## Contribution rule

Add a scenario only when it corresponds to:
1. a real implementation behavior or failure;
2. a deterministic evidence path;
3. a consequential sink;
4. a portable authority/closure distinction.

Do not grow this into an abstract taxonomy.

## Current evidence status

- portable invariant: established across multiple runtimes;
- reference evaluator: this directory;
- real third-party adapter: not yet;
- independent reuse: not yet;
- standards adoption: not yet.
