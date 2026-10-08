# Standalone Extraction Boundary

Status: temporary staging boundary
Date: 2026-10-08

This directory is **not a ClawManager product feature**.

ClawManager is a forked Kubernetes-first agent/runtime control-plane project. The files under this directory are separate experimental work on authorization-closure conformance and are staged here only because an independent repository did not yet exist when the work was developed.

## Intended identity

This component should eventually live in an independent repository with its own:
- repository history;
- README;
- issue tracker;
- CI;
- contribution policy;
- release/versioning;
- adapters and conformance fixtures.

Do not describe ClawManager itself as an Authorization Closure or Runtime Accountability project.

## Extraction root

The standalone component is exactly:

- conformance/authorization_closure/

Current functional files include:
- README.md
- runner.py
- test_runner.py
- fixtures/
- adapters/

The GitHub Actions workflow currently exercising this component lives at:
- .github/workflows/authorization-closure-conformance.yml

When extracted, copy that workflow into the new repository and retarget its paths to the repository root.

## Runtime dependencies

The evaluator and current adapters use only the Python standard library.

No ClawManager application package, API, database schema, Kubernetes object, frontend, or deployment asset is required.

This zero-dependency boundary is intentional: it allows the conformance layer to be extracted without carrying ClawManager implementation baggage.

## Evidence provenance

The component consumes **evidence about** independent systems. It does not copy their runtime implementations.

Current evidence adapters:

### OpenClaw

Maps observable facts established by a merged OpenClaw regression around execution admission revocation before native process/PTY construction.

This is evidence normalization, not OpenClaw code ownership and not OpenClaw adoption of this harness.

### Open Agent Auth

Maps upstream revocation-test evidence and current Resource Server enforcement uncertainty.

The correct current result is UNKNOWN when sink-side closure evidence is absent.

This is evidence normalization, not Alibaba adoption of this harness.

## Attribution discipline

After extraction, preserve links back to:
- the upstream OpenClaw regression used as evidence;
- the Open Agent Auth files/tests used as evidence;
- any standards drafts used for terminology or comparison.

Do not claim:
- authorship of upstream regressions;
- invention of revocation-closure terminology already present in standards work;
- external adoption where only local adapters exist.

## Migration trigger

Extract this component into an independent repository as soon as repository creation is available.

After successful extraction and verification:
1. preserve this file or a short tombstone in ClawManager;
2. remove the executable conformance subtree from ClawManager;
3. point historical evidence notes to the independent repository;
4. leave ClawManager focused on its actual control-plane purpose.

## Identity goal

The independent project should earn its identity through:

real implementation evidence
-> executable conformance
-> independent adapters
-> external reuse
-> standards/conformance participation

not through association with the ClawManager fork.
