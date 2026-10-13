# Revocation Closure — Implementer Conformance Matrix v0.1

Contributor: Noah Ellis (@hippoley)
Target: draft-watts-oauth-agent-revocation-closure-01, sections 5–7 and 11–12.
Status: proposed implementation-neutral test specification. NOT executed, adopted, or endorsed.

## Why this is additive

The draft -01 already includes six illustrative conformance and falsification cases in Section 12. We do not claim these concepts are new. The contribution is to define concrete experiment inputs, observable Resource Server and sink behavior, evidence requirements, and independent failure conditions. Vendor-specific security disclosures are intentionally excluded.

## Minimum test adapter

Required operations: issue(epoch, principal, scope), authorize(token, sink), revoke(epoch), dispatch(work, token, sink), commit(work), observe(sink), snapshot_authority(). Implementations may map these to their native interfaces; this adapter does not define a new OAuth protocol.

Required evidence per run: run ID, implementation/adapter version, AS/RS identity, authority epoch, declared scope, consequential sink, closure budget and effective boundary, authoritative event ordering or sequence numbers, revocation acknowledgement, RS authorization decision, sink-side commit ledger, original input/output digests, unresolved paths, and evidence provenance. Wall-clock timestamps alone are not proof of commit ordering.

## Six cases mapped to -01

| ID | Draft mapping | Trigger | Expected observations | Falsification / evidence |
|---|---|---|---|---|
| RC-01 | §12 case 1 | Issue T and unrevoked U; confirm both accepted; revoke T epoch; retry after effective boundary | T cannot commit a new in-scope effect; U remains subject to normal authorization | A T-derived post-boundary sink commit invalidates CLOSED. Preserve AS response, RS decision and sink ledger. |
| RC-02 | §12 case 1, split service | AS and RS in distinct processes; RS caches pre-revoke signature verification; AS revokes | RS enforcement reflects revocation by its declared propagation deadline | AS ACK with T-derived post-deadline in-scope commit invalidates CLOSED; AS-only deny list is not enough. |
| RC-03 | §12 case 3 | Make an in-scope queue or RS revocation state inaccessible | Report PARTIAL/UNKNOWN without authenticated sink-side enforcement coverage | CLOSED inferred solely from AS ACK or missing telemetry is false evidence. |
| RC-04 | §12 case 2; §11 | Queue a permitted job, revoke authority, then drain after alleged closure | No new E-attributed in-scope sink commit after closure | A post-closure commit disproves CLOSED; a cancellation request alone is insufficient. |
| RC-05 | §12 cases 4–5 | Replay E1 closure receipt for E2, change scope, or use stale evidence | Independent verifier rejects cross-epoch or stale claims; E2 reauthorized effects not misattributed | Acceptance of wrong-epoch receipt invalidates verification. |
| RC-06 | §12 case 6 | Commit before closure and observe persistent effect afterward | Old effect remains without retroactively falsifying closure; no new E effect commits | Confusing observation time with commit time produces incorrect verdict. |

## Engineering acceptance gates

1. Static source analysis: observations are HYPOTHESES, not established production vulnerabilities.
2. Reference simulator: fail deliberately broken implementation and pass control implementation, with test logs labelled SIMULATED.
3. Independent split-process AS/RS and consequential sink: collect real execution records and declared propagation assumptions.
4. Repeat on a second implementation: capture reproducible independent feedback, with no conflation of an ACK and closure.
5. Request author review of compatibility with non-normative §12 and separately discuss attribution before submission.

## Questions to Deonté

- Would an implementation-neutral runner + evidence schema be useful as a companion to existing non-normative §12?
- Which closure-budget and commit-time fencing conditions should be pinned down first?
- What is the minimum sufficient evidence to distinguish scoped CLOSED from PARTIAL and UNKNOWN?
- Can we agree on explicit contributor attribution for reviewed test design, code, and verified implementation results before incorporation?

## Safety and publication

Do not publish embargoed vendor-specific details, secrets, production endpoints, exploit steps, or unverified claims of a named implementation vulnerability. Send any sensitive evidence only through the vendor's coordinated disclosure channel. This matrix is not a certificate and does not confer IETF approval.
