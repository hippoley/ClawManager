# OpenClaw #166495 — remaining Linux Doctor + published-driver validation contract

Target head: `02c316342775562101fc980b3728fab24e871c3e`

## What is already proven

The lower publication boundary is already covered by merged fs-safe #843:

- AWS/Linux selective-seccomp run;
- real native addon;
- `renameat2(RENAME_NOREPLACE)` denied;
- auto-mode link/unlink fallback succeeds for payload + sibling manifest publication;
- collision, source substitution, link denial, failed unlink, and strict require-mode cases covered;
- full fs-safe validation passed.

Therefore #166495 does **not** need to re-prove the syscall implementation.

The remaining evidence should prove the OpenClaw consumer/entrypoint composition.

## 1. Linux Doctor/seccomp boundary

Goal:

```text
real OpenClaw Doctor/update capture owner
→ fs-safe 0.24.2 native auto publication
→ interrupted/partial state remains evidence
→ discovery/retirement consume it conservatively
```

### Minimum useful run

Use a Linux host/container where the same selective seccomp rule denies
`renameat2(RENAME_NOREPLACE)` while normal link/unlink remain available.

Run a real Doctor repair flow that creates an update-recovery baseline capture.

The proof should record:

```text
platform = linux
fs-safe version = 0.24.2
rename-noreplace = denied by the selective filter
Doctor/update capture completes through native auto fallback
manifest.json exists
manifest.json.partial is absent after successful publication
manifest nlink = 1
Doctor command exits successfully
```

Then run the focused consumer checks on the same candidate:

```bash
pnpm test   src/infra/update-database-backup.test.ts   src/cli/update-cli/status.recovery.test.ts   --maxWorkers=1
```

The runtime boundary evidence and consumer regressions prove different layers; keep both.

### Interrupted-state control

Do not try to crash the real Doctor at an uncontrolled instant.

Use the deterministic filesystem regressions already in the PR for:

- linked final + partial;
- copied final + partial;
- contradictory partial contents.

The live Linux run only needs to prove that OpenClaw reaches fs-safe's fallback successfully from the
real Doctor/update capture entrypoint.

## 2. Published-driver / candidate compatibility

The repository already owns this contract through the published-upgrade survivor harness.

Use a pinned published predecessor and a source-pinned candidate artifact.

A minimal existing lane is:

```bash
OPENCLAW_UPGRADE_SURVIVOR_BASELINE_SPEC=openclaw@2026.8.33 \
OPENCLAW_UPGRADE_SURVIVOR_SCENARIO=legacy-operator-state \
pnpm test:docker:published-upgrade-survivor
```

Why this is useful:

- the **published baseline CLI** drives the update;
- the harness verifies installed candidate payload identity, not only version strings;
- the published driver must update/restart the managed Gateway in the normal auto-auth release lane;
- candidate Doctor/config/state checks run after the update;
- assertions occur before standalone Doctor can hide an incomplete migration.

For a narrower candidate-byte identity check, the repository docs also allow source-pinned
`base` / `sqlite-volume` survivor runs; they compare the installed application payload with the
frozen candidate tarball after update.

### Acceptance record

Capture:

```text
Published driver: <exact openclaw@version>
Candidate commit: 02c316342775562101fc980b3728fab24e871c3e
Candidate tarball digest: <sha256>
Scenario: legacy-operator-state (or selected narrower supported lane)

published updater exit: 0
installed candidate identity: exact match
Gateway restart/readiness: success
Doctor/core validation: success
no recovery-capture compatibility refusal
```

## 3. Candidate/prepared generation compatibility

#166495 strengthens `readBinding()` by requiring a complete seal before treating candidate/prepared
generations as verified rollback points.

The focused PR tests already cover:

```text
partial manifest remains
→ incompleteGenerations fingerprint only
→ no candidateSha256/preparedSha256 authority granted

complete sealed generation
→ existing run/baseline/candidate identity checks still apply
```

For merge evidence, record those focused tests together with the published-driver run. A separate
historical runtime reproduction is unnecessary unless the survivor lane exposes a failure.

## 4. What not to claim

Do not claim:

- QNAP hardware proof;
- arbitrary Linux filesystem coverage;
- crash atomicity of link/unlink publication;
- automatic repair/removal of retained final+partial pairs.

The supported claim is narrower:

> On the tested Linux seccomp boundary, real OpenClaw capture publication reaches fs-safe 0.24.2's
> supported fallback, and #166495 conservatively classifies interrupted publication evidence while
> preserving published-driver/candidate compatibility.

## Suggested PR-body evidence block

```text
Linux Doctor/seccomp:
- host: AWS/Linux (redacted instance identity)
- fs-safe: 0.24.2
- selective seccomp denied rename-noreplace
- real OpenClaw Doctor/update capture completed
- final manifest present, partial absent, nlink=1
- focused recovery suites: PASS

Published-driver/candidate:
- baseline: openclaw@2026.8.33
- candidate: 02c316342775562101fc980b3728fab24e871c3e
- scenario: legacy-operator-state
- published updater: PASS
- candidate payload identity: PASS
- managed Gateway replacement/readiness: PASS
- Doctor/core validation: PASS
```

Status: source-reviewed against #166495 and repository testing documentation; not executed by
`hippoley`.
