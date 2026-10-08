# Temporal — Principal Attribution vs execution authorization continuity

Status: source/docs-reviewed contract note
Date: 2026-10-08

## Existing Temporal primitives

Temporal currently exposes three adjacent but distinct mechanisms:

1. **Authorizer / ClaimMapper**
   - controls whether incoming Temporal API calls are allowed;
   - operates at the Temporal Frontend/API boundary.

2. **Principal Attribution**
   - when enabled, stamps the Principal returned by the Authorizer onto Event History events;
   - provides auditability for who initiated/caused API-visible history changes.

3. **Context Propagation / Interceptors**
   - can carry application metadata such as user/tenant/auth tokens across Client -> Workflow -> Activity / Child Workflow boundaries;
   - can also apply cross-cutting authorization/header logic in SDK code.

These mechanisms are complementary, but they do not automatically establish one end-to-end authority contract for a side-effecting Activity.

## Core distinction

> **Attribution != propagated identity != current execution authorization.**

A Principal in Event History proves an attributed caller for a Temporal API/history event.

A propagated user/tenant value proves application metadata was carried across SDK boundaries.

Neither fact alone proves that the same principal is still authorized at the moment a long-running/retried Activity performs an external side effect.

## Durable-execution question

Consider:

```
user U starts workflow at t1
-> Event History is attributed to U
-> user/tenant context is propagated
-> workflow waits/retries for hours
-> U loses permission at t2
-> Activity attempt executes external side effect at t3
```

Potentially relevant facts are now distinct:

```
historical initiating principal = U
propagated application identity = U
current authorization at t3 = ?
activity attempt generation = gN
external side-effect truth = ?
```

A durable runtime may legitimately choose different policies, but the contract should not collapse these facts.

## Why this matters for autonomous agents

Agent runtimes increasingly combine:
- long-lived workflow state;
- durable retries/replay;
- delegated tools;
- human approvals;
- identity/policy changes;
- irreversible side effects.

As agents become more autonomous, "who originally initiated this?" and "who may still authorize this action now?" diverge more often.

## Suggested documentation clarification

A small docs clarification could state explicitly that:

- Principal Attribution is audit attribution for Temporal history/API operations;
- application identity propagated through headers/context is application-defined metadata;
- authorization of business/tool side effects at Activity execution time remains an application/security-layer responsibility unless a stronger product contract is documented.

This would help prevent users from assuming that an attributed principal automatically implies end-to-end business authorization continuity.

## Suggested future conformance scenario

If Temporal/Oso introduces an agent-security identity/policy surface, one useful scenario is:

```
initiating principal remains visible in history
+
application identity is propagated
+
authority is revoked before side-effecting Activity attempt
        ↓
historical attribution remains intact
but
current dispatch/side-effect authorization follows the new policy
```

That scenario should only become a runtime test when an owning authorization API exists.

## Non-claim

This note does not claim a Temporal vulnerability or missing security control.

It only separates the current documented mechanisms so future agent-security semantics can compose them without ambiguity.
