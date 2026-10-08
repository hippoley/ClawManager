# OpenClaw runtime-correctness credibility ledger

Last refreshed: 2026-10-08

This is a provenance index, not a claim of maintainership. It records externally verifiable cases where hippoley's analysis, issue definition, or proof design was independently reviewed, adopted, or merged upstream.

## 1. PR #166503 — custody vs authority

Upstream:
https://github.com/openclaw/openclaw/pull/166503

External signal:
- ClawSweeper explicitly said its resolved custody finding agreed with hippoley's source analysis.
- Maintainer repeatedly re-ran exact review.
- PR ultimately squash-merged as commit `cb01febf9fa4fbda271d10b3adff3b384c6ece76`.

Durable capability evidenced:
- distinguish physical reader custody from execution authority;
- identify the exact predicate that should survive post-cleanup;
- reason at the source owner boundary rather than from surface state.

Reusable invariant:

> Custody settlement is not authority revocation.

## 2. Issue #166466 → PR #166495 — incomplete publication evidence

Canonical issue:
https://github.com/openclaw/openclaw/issues/166466

Maintainer implementation:
https://github.com/openclaw/openclaw/pull/166495

External signal:
- #166466 formalized the consumer gap around retained partial publication evidence.
- Maintainer-authored #166495 explicitly closed #166466.
- PR squash-merged as `a86dd383534f8b3b6d1ebdc7e4b7601db65c6ad3`.

Durable capability evidenced:
- detect second-order consumer defects after a lower-level publication mechanism changes;
- distinguish a final-looking artifact from authoritative completion;
- reduce the repair to conservative evidence classification.

Reusable invariant:

> Observable final-looking state is not authoritative completion.

## 3. PR #154728 — third-party proof-path reuse

Upstream:
https://github.com/openclaw/openclaw/pull/154728

External signal:
- hippoley pointed the author to the existing Gateway handler harness.
- author `gokay-ai` added the regression and wrote: “Thanks @hippoley for pointing at that harness.”
- hippoley then identified the stronger `startGatewayWithClient()` / real WebSocket GatewayClient seam.
- author adopted it and again wrote: “Thanks @hippoley for the pointer.”

Durable capability evidenced:
- find the production proof seam that satisfies a review blocker;
- separate helper-level evidence from real transport evidence;
- improve another contributor's merge-readiness without taking over their PR.

Reusable invariant:

> Proof should cross the same ownership and transport boundary as the behavior being claimed.

## 4. Issue #166770 — current ownership position

Upstream:
https://github.com/openclaw/openclaw/issues/166770

Public repair contract:
https://github.com/hippoley/ClawManager/blob/main/docs/patches/openclaw-166770-semantic-progress-from-settled-tool-execution.md

Candidate implementation boundary:
https://github.com/hippoley/ClawManager/blob/main/docs/patches/openclaw-166770-candidate-implementation-boundary.md

Current status:
- canonical P1 issue;
- source repro / fix-shape-clear / queueable-fix;
- no fixing PR found at the time of intervention;
- automatic implementation stopped before deterministic gates completed;
- hippoley posted a bounded owner/settlement repair contract;
- later source review corrected a worker-only coverage mistake and aligned the candidate repair with the repository's existing event-object + owner-generation provenance pattern;
- a separate candidate implementation-boundary artifact now freezes the proven parts and explicitly leaves only the host-private embedded-owner injection seam unresolved;
- no third-party adoption or routing of #166770 had appeared at the latest refresh.

Durable capability being tested:
- distinguish model response shape from actual execution truth;
- bind semantic progress to successful settlement under the exact current owner;
- preserve failed/blocked/stale-owner negatives.

Reusable invariant:

> Semantic progress follows validated execution, not the enclosing response's terminal shape.

## Identity progression evidenced so far

```
source-level distinction
  → independent review attribution
  → maintainer adoption + merged repair
  → independent contributor reuse + explicit credit
  → current attempt to own a new bounded correctness boundary
```

## What is not yet proven

Do not overclaim:

- no established maintainer role;
- no recurring maintainer routing to @hippoley yet;
- no clear upstream PR authored by hippoley and carried personally through review to merge in this sequence yet;
- #154728 is still open at this refresh;
- #166770 is a current position, not a landed precedent yet.

The next durable credential is therefore a self-authored, bounded upstream correctness repair that survives proof, review, revision, and merge.
