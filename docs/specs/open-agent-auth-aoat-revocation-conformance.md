Open Agent Auth conformance case — AOAT revocation continuity

Date: 2026-10-08
Status: source-reviewed executable-contract draft
Target: alibaba/open-agent-auth

Related project direction:
- Open Agent Auth issue #14 proposed a layered Protocol Conformance Test Suite as a living executable specification.
- Current docs claim AOAT revocation, immediate invalidation, and per-use validation.

This case defines the smallest cross-service behavior needed for those claims to be true independent of implementation strategy.

## OAA-RC-1 — Revoked AOAT must be rejected before expiry

Given:
- an Authorization Server (AS);
- a Resource Server (RS);
- an AOAT T issued by AS;
- T is cryptographically valid;
- T is not expired;
- T is accepted by RS before revocation.

When:
- AS revokes T using the configured revocation mechanism;
- no token claims or signature are modified;
- T is presented to RS again before expiration.

Then:
RS MUST reject T as no longer currently authorized.

Formal property:
valid_signature(T) && now < exp(T) && revoked(T) => reject_resource_access(T)

The conformance property is about current authorization, not token syntax.

## OAA-RC-2 — Non-revoked control remains valid

Given an otherwise equivalent token U that has not been revoked:
valid_signature(U) && now < exp(U) && !revoked(U) => normal authorization pipeline continues

This control prevents a repair from accidentally disabling all AOATs.

## OAA-RC-3 — Revocation survives validator/retry reuse

Repeat OAA-RC-1 using the same Resource Server process and the same validator instance that previously accepted T.

Expected:
- prior successful validation must not pin T as authorized;
- cached signature/parsing work may be reused;
- cached authorization validity must not bypass fresh revocation state.

Portable invariant: cacheable cryptographic validity is not cacheable current authority.

## OAA-RC-4 — Split-service revocation propagation

Run AS and RS as separate service instances.
1. RS accepts T.
2. AS records revocation of T.
3. RS receives T again before expiry.

Expected: RS rejects T once the documented revocation propagation contract says the revocation is effective.

Acceptable implementations include:
- shared/distributed revocation storage;
- RFC 7662 introspection;
- pushed invalidation;
- another explicit authoritative revocation backend.

A local in-memory blacklist on AS alone does not satisfy this distributed case.

## OAA-RC-5 — Resource Server unavailable to revocation authority

If the RS cannot determine current revocation status when the product contract requires immediate revocation, the implementation must define a policy.

Possible policies:
- fail closed for security-sensitive operations;
- allow within a documented bounded staleness window;
- operation-class-specific behavior.

The important requirement is that the behavior is explicit and testable.
Do not silently treat revocation status unavailable as not revoked unless that weaker availability-over-revocation contract is explicitly documented.

## Current source evidence

At the reviewed default-branch snapshot:
- TokenRevocationService exposes revoke(token) and isRevoked(token);
- InMemoryTokenRevocationService implements that state locally;
- AoatValidator.validate() verifies signature, expiration, issuer, audience, and required claims;
- OperationAuthorizationValidator delegates to AoatValidator;
- DefaultResourceServer builds the five-layer verifier without a revocation service;
- repository search found no RS/AOAT acceptance path consuming isRevoked(...);
- no Resource Server token-introspection client / remote revocation-state consumer was found.

This is why the distributed contract matters more than merely injecting one local service into one validator.

## Minimal implementation progression

Phase 1 — prove the current gap
Add a core/unit regression where a validator explicitly connected to a revocation authority rejects a revoked AOAT.

Phase 2 — wire the Resource Server
Ensure RS acceptance consults current revocation authority.

Phase 3 — prove split-service behavior
Use integration tests with distinct AS and RS instances. Revocation at AS must become observable at RS through the selected production mechanism.

Phase 4 — cache safety
If token validation results are cached, separately classify immutable/slow-changing cryptographic facts and mutable authority facts such as revocation. Only the former may be safely cached for the full remaining token lifetime without invalidation.

## Why this is a conformance issue, not one implementation detail

The external behavior remains the same whether implementation uses TokenRevocationService, RFC 7662 introspection, a distributed deny set, or push-based invalidation.

The durable test should assert: a revoked AOAT cannot authorize a new protected operation after revocation becomes effective.

## Cross-runtime significance

This is a concrete instance of the broader Durable Authority rule:
Historical authorization artifact != current execution authority.

The same distinction appears in agent operation tokens, durable workflow approvals, delegated credentials, long-lived MCP sessions, cloud control-plane agents, payments, and irreversible external side effects.

## Current evidence level

project docs/security contract observed = YES
source acceptance path traced = YES
gap reported privately = YES
maintainer confirmation = NOT YET
conformance test landed = NOT YET
fix landed = NOT YET
public credit = NOT YET

Do not overclaim beyond this level.
