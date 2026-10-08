# Open Agent Auth — AOAT revocation continuity gap

Date: 2026-10-08
Status: source-reviewed implementation gap, not executed against a running deployment

Target:
- alibaba/open-agent-auth

## Public contract

Current docs state all of the following:

1. AOAT/JWT `jti` supports token revocation tracking.
2. Token blacklists provide immediate revocation for security incidents.
3. MCP/resource-server validation should reject revoked tokens before expiration.
4. AOATs may be cached, but must still be validated on each use so revoked/expired tokens are not accepted.

## Current source path

### Revocation service exists

`TokenRevocationService` exposes:

```java
void revoke(String token);
boolean isRevoked(String token);
```

with `InMemoryTokenRevocationService` implementing the revoked-token set.

### AOAT validation does not currently consume it

`AoatValidator.validate(String aoatJwt)` currently verifies:

- signature;
- expiration;
- issuer;
- audience;
- required claims.

It does not receive or call `TokenRevocationService`.

### Layer 3 delegates directly to AOAT validator

`OperationAuthorizationValidator.validate(...)` calls:

```java
aoatValidator.validate(aoatJwtString)
```

and then optional delegation-chain validation.

No revocation check is visible on this path.

### Resource Server uses that validator in the five-layer pipeline

`DefaultResourceServer` builds the five-layer verifier with:
- WIT validator;
- WPT validator;
- AOAT validator;
- policy evaluator;
- binding instance store.

No token revocation service is passed into the Resource Server verification pipeline.

### Search result

At the reviewed source snapshot, `TokenRevocationService.isRevoked(...)` is referenced only by:
- the interface;
- its in-memory implementation;
- its unit tests.

No Resource Server / AOAT validation consumer was found.

## Durable Authority invariant

> Revocation is an authority-state transition, not merely a token-management side effect.

A token that was valid at issuance but revoked before a later use should not retain execution authority merely because its signature and expiration remain valid.

```
cryptographically valid
+
not expired
!=
currently authorized
```

when revocation is part of the documented contract.

## Minimal regression shape

1. issue/generate an AOAT that validates successfully;
2. assert resource-server validation succeeds;
3. revoke that exact token through the configured `TokenRevocationService`;
4. validate the same AOAT again before expiration;
5. expect resource-server validation to fail.

Negative control:
- an otherwise identical non-revoked AOAT continues to validate.

## Minimal repair families

Any of these can satisfy the contract:

1. inject `TokenRevocationService` into `AoatValidator` and check before success;
2. add a dedicated revocation validator layer before AOAT acceptance;
3. integrate RFC 7662 introspection / shared revocation state at the resource-server boundary.

The key requirement is semantic:

> every accepted resource request must establish current non-revoked authority, not just historical token validity.

## Non-claims

This note does not establish:
- a remotely exploitable deployment;
- that every deployment uses the in-memory revocation service;
- that distributed revocation propagation is already specified;
- that the docs are not intentionally ahead of implementation.

It establishes only a current docs/source mismatch worth clarifying and regression-testing.
