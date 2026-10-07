# OpenClaw #155320 — current-head test-cost handoff

Target head: `82dfc68daa49665c21b7426704f43a939e9c5e98`

Current ClawSweeper Revision 9 has one remaining merge item:

> Update the PR evidence with measured focused before/after wall time for the consolidated head and CI test seconds when available; the current 159-second, four-test result describes the earlier layout.

## What is already verifiable

### Current-head CI

GitHub Actions workflow for this exact head:

- CI run: `36365670595`
- conclusion: **success**
- current-head CI gate status: **success**
- gate run: `36366863675`

The exposed jobs for run `36365670595` show successful preflight / security / contract / plan lanes.

The public first-page job list does **not** expose a Node/Gateway test shard that can be safely attributed to:

- `src/gateway/server.subagent-prompt-recent.gateway.test.ts`
- `src/gateway/server.subagent-prompt-recent.gateway.test-support.ts`

Therefore do **not** report a check-plan or whole-workflow duration as per-file CI seconds.

## What still needs one local measurement

The current consolidated head shares one Gateway across the recoverable and settled phases, so the earlier:

```text
159 seconds / four tests
```

must not be reused as if it were the current layout.

Run the exact focused file on the current head and record wrapper wall time plus Vitest duration:

```bash
/usr/bin/time -p pnpm test src/gateway/server.subagent-prompt-recent.gateway.test.ts --maxWorkers=1
```

If the support module is imported by that file, do not run it separately merely to inflate or duplicate timing.

If a second directly owned unit file is part of the current PR's focused set, run it separately and report it separately.

## Recommended PR-body wording

```md
### Current-head test cost

Head: `82dfc68daa49665c21b7426704f43a939e9c5e98`.

Focused consolidated Gateway regression:

- command: `pnpm test src/gateway/server.subagent-prompt-recent.gateway.test.ts --maxWorkers=1`
- result: <N passed>
- wrapper wall: <X.XX s>
- Vitest duration: <Y.YY s>

Current-head GitHub Actions CI run `36365670595` completed successfully, and the repository CI gate for this head is green via run `36366863675`.

The exposed Actions job list does not provide a safely attributable per-file CI duration for this Gateway regression, so no per-file CI seconds are claimed. Shared job/workflow time is intentionally not relabeled as file timing.
```

## Why this should satisfy the review ask

This replaces the stale 159-second/four-test timing with a measurement from the actual consolidated head, while keeping CI provenance honest:

```text
focused local wall time
→ current exact head

CI success
→ current exact head

per-file CI seconds
→ unavailable from exposed job data
→ explicitly not claimed
```

Status: GitHub CI metadata verified for the exact head; no local test execution is claimed in this artifact.


## Self-audit (2026-10-07)

The focused timing command matches OpenClaw's own test-cost policy:

- repository `AGENTS.md` requires PRs to report `pnpm test <file> --maxWorkers=1` wall time and CI seconds when available;
- `docs/help/testing/writing-tests.md` defines the same one-worker command as the cost-budget measurement;
- the test-performance skill also uses `/usr/bin/time ... pnpm test <file> --maxWorkers=1`.

So the proposed consolidated-head measurement command follows the repository's documented maintenance contract rather than inventing a custom runner shape.
