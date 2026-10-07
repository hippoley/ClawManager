# hippoley

Building agent infrastructure, evaluation systems, spatial intelligence, and evidence-first open-source tooling.

## Current focus

I work on systems where AI agents have to operate under real runtime constraints: state, tools, memory, provenance, physical environments, and human approval.

Selected projects:

- [CounterProof](https://github.com/hippoley/CounterProof) — provenance, lifecycle verification, and evidence-oriented agent workflows.
- [ContextMesh](https://github.com/hippoley/ContextMesh) — context retrieval and degradation measurement.
- [MeshFit](https://github.com/hippoley/MeshFit) — heterogeneous model placement and execution benchmarking.
- [HumanQueue](https://github.com/hippoley/HumanQueue) — human-in-the-loop coordination for waiting agents and sessions.
- [AirTrajectory](https://github.com/hippoley/AirTrajectory) — physical airflow / control orchestration experiments.
- [SpatialRuntime](https://github.com/hippoley/SpatialRuntime) — spatial interaction runtime experiments.

## Upstream work

I prefer evidence over contribution-count vanity. Upstream work is classified by public provenance:

- **Authored** — directly attributable to `hippoley`.
- **Contributed evidence** — public reproduction, benchmark, compatibility proof, review, or runtime evidence from `hippoley`.
- **Upstream incorporated** — related work was superseded or reimplemented upstream; not represented as my merged PR unless GitHub attribution proves it.
- **Tracked upstream** — engineering context only, with no authorship claim.

### OpenClaw

Active upstream work is tracked in the [provenance-first OpenClaw engagement ledger](https://github.com/hippoley/ClawManager/blob/main/docs/upstream-openclaw-engagement.md).

The objective is simple: turn blocker-specific engineering work into public, maintainer-usable artifacts — focused PRs, runtime proofs, compatibility evidence, concrete reviews, and reproducible issues.

## Engineering style

- Reality before narrative.
- Reproducible evidence before claims.
- Small bounded patches before framework expansion.
- Explicit compatibility and migration boundaries.
- Runtime proof where behavior matters.
- Preserve attribution when upstream incorporates or supersedes work.

## What I am looking for

Early infrastructure problems where a small amount of rigorous engineering can become a durable protocol, standard, runtime primitive, or open-source building block.
