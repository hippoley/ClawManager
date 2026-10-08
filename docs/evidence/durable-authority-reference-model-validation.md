# Durable Authority reference model validation

Date: 2026-10-08

Artifact:
- `tools/durable_authority_reference_model.py`

Purpose:
Validate that the initial Durable Authority Continuity scenarios can be represented as an executable state model rather than prose-only distinctions.

Executed scenarios:

1. approval survives crash / unchanged authority;
2. historical approval does not override a later authority revision;
3. UNKNOWN outcome blocks redispatch until reconciliation;
4. ABSENT prior effect does not recreate revoked authority;
5. stale completion cannot overwrite a successor execution generation;
6. yielded continuation preserves exact obligation custody.

Observed result:

```
6/6 reference scenarios passed
```

Important limitation:

This validates only the internal consistency of the reference model.

It does **not** establish that:
- Temporal Agent Harness implements these semantics;
- OpenClaw implements every scenario;
- LangGraph implements every scenario;
- DAC v0.1 is an accepted external standard.

The next evidence step must come from a real runtime adapter/probe or maintainer response.

Why this matters:

The research line has moved from:

```
prose invariant
-> machine-readable vectors
-> executable reference state model
```

The next legitimate upgrade is:

```
reference state model
-> framework adapter
-> real runtime observation
-> external adoption / merged regression
```

Do not advance the identity claim beyond that evidence.
