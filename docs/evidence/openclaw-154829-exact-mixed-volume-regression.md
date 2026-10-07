# OpenClaw #154829 — exact current-main mixed-volume regression

This regression targets the current split Doctor lint runner and proves the actual boundary ClawSweeper is asking about:

- operator state lives on one logical volume;
- selected plugin inspection runs against its private snapshot;
- `core/doctor/disk-space` must still probe the operator/source state directory.

## Production patch prerequisite

Current-main production change:

```ts
// src/commands/doctor-disk-space.ts
function collectDiskSpaceWarnings(params?: { env?: NodeJS.ProcessEnv }) {
  const env = params?.env ?? process.env;
  const homedir = () => resolveRequiredHomeDir(env, os.homedir);
  const stateDir = resolveStateDir(env, homedir);
  // ...
}

export function collectDiskSpaceHealthFindings(params?: {
  env?: NodeJS.ProcessEnv;
}): readonly HealthFinding[] {
  const result = collectDiskSpaceWarnings(params);
  // ...
}
```

and:

```ts
// src/flows/doctor-health-contributions-initial.ts
async detect(ctx) {
  const { collectDiskSpaceHealthFindings } =
    await import("../commands/doctor-disk-space.js");
  return collectDiskSpaceHealthFindings({ env: ctx.env });
}
```

No-argument callers such as `noteDiskSpace()` remain unchanged.

## Unit regression for explicit environment

Add to `src/commands/doctor-disk-space.test.ts`:

```ts
it("uses an explicit source environment instead of ambient state", () => {
  vi.stubEnv("OPENCLAW_STATE_DIR", "/private/plugin-snapshot");

  const tryReadDiskSpace = vi
    .spyOn(diskSpace, "tryReadDiskSpace")
    .mockReturnValue({
      availableBytes: 499 * 1024 * 1024,
      targetPath: "/operator/state",
      checkedPath: "/operator",
      totalBytes: null,
    });

  const findings = collectDiskSpaceHealthFindings({
    env: {
      ...process.env,
      OPENCLAW_STATE_DIR: "/operator/state",
    },
  });

  expect(tryReadDiskSpace).toHaveBeenCalledWith("/operator/state");
  expect(findings).toEqual([
    expect.objectContaining({
      checkId: "core/doctor/disk-space",
      path: "/operator/state",
      requirement: "low-free-space",
    }),
  ]);
});
```

This directly proves the collector no longer depends on ambient snapshot state when the health-check owner supplies `ctx.env`.

## Mixed selected-check integration regression

In `src/commands/doctor-lint.test.ts`, extend the existing
`keeps mixed selected checks on an isolated plugin metadata view` case.

Add a `tryReadDiskSpace` spy and select disk-space alongside the plugin check:

```ts
const diskProbePaths: string[] = [];
const tryReadDiskSpace = vi
  .spyOn(diskSpace, "tryReadDiskSpace")
  .mockImplementation((pathname) => {
    diskProbePaths.push(pathname);
    return {
      availableBytes: 10 * 1024 * 1024 * 1024,
      targetPath: pathname,
      checkedPath: pathname,
      totalBytes: null,
    };
  });

await expect(
  runDoctorLintCli(runtime, {
    json: true,
    severityMin: "error",
    onlyIds: [
      "memory-core/managed-local-embedding-setup",
      "test/source-config-interpolation",
      "core/doctor/disk-space",
    ],
  }),
).resolves.toBe(0);

expect(JSON.parse(String(stdout.mock.calls.at(-1)?.[0]))).toMatchObject({
  ok: true,
  checksRun: 3,
  findings: [],
});

expect(inspectSourceConfig).toHaveBeenCalledOnce();

// Critical assertion: disk capacity follows the operator/source environment.
expect(diskProbePaths).toContain(stateDir);

// Existing isolation assertions remain unchanged.
expect(mocks.readConfigFileSnapshot).not.toHaveBeenCalled();
expect(sourceOpenStacks).toEqual([]);
expect(snapshotDoctorLintSqliteFamily(databasePath)).toEqual(before);
```

Import the disk-space owner in that test module if it is not already present:

```ts
import * as diskSpace from "../infra/disk-space.js";
```

## What this proves

The combined regression demonstrates all three contracts simultaneously:

```text
source env
  └─ disk-space check -> operator stateDir

private plugin snapshot
  └─ plugin metadata / semantic inspection

source SQLite family
  └─ unchanged
```

That is stronger than a direct collector unit test alone because it exercises the registered Doctor contribution through the real selected-check path.

## Compatibility declaration

This patch changes no:

- persisted schema;
- configuration key;
- migration/backfill owner;
- SQLite content;
- plugin snapshot isolation policy.

It only makes the disk-capacity dependency explicit at the registered health-check boundary.

## Focused validation

```bash
node scripts/run-vitest.mjs src/commands/doctor-disk-space.test.ts --run
node scripts/run-vitest.mjs src/commands/doctor-lint.test.ts --run
```

For the published-driver compatibility cell already present in the PR, rerun it on the refreshed candidate and report its exact head SHA separately.

Status: source-reviewed against current main on 2026-10-07; not executed in this environment.
