# OpenClaw #154735 — exact current-main deep-status regression

Target: current `main` after reusing the shared formatter owners that have already landed.

Add this focused regression to `src/commands/status.command-report-data.test.ts`:

```ts
it("shows context-engine quarantine and exhausted config reload in deep status", async () => {
  const params = createStatusCommandReportDataParams();
  const health = expectDefined(params.health, "health fixture");

  health.contextEngines = {
    quarantined: [
      {
        engineId: "lossless-claw",
        owner: "plugin:lossless-claw",
        operation: "assemble",
        reason: "database corrupt",
        failedAt: Date.now(),
      },
    ],
  };
  health.configReload = { hotReloadStatus: "disabled" };

  const report = await buildStatusCommandReportData({
    ...params,
    opts: { deep: true },
    health,
  });

  const rows = report.healthRows?.map(({ Item, Status, Detail }) => ({
    Item,
    Status: stripAnsi(Status),
    Detail,
  }));

  expect(rows).toEqual(
    expect.arrayContaining([
      {
        Item: "Context engine",
        Status: "WARN",
        Detail: "warning (1 quarantined; downgraded to legacy: lossless-claw)",
      },
      {
        Item: "Config hot reload",
        Status: "WARN",
        Detail: "disabled (watcher retries exhausted; restart the gateway to restore it)",
      },
    ]),
  );
});
```

And keep/add this healthy omission case:

```ts
it("omits auxiliary operational rows when healthy", async () => {
  const params = createStatusCommandReportDataParams();
  const health = expectDefined(params.health, "health fixture");

  health.contextEngines = { quarantined: [] };
  health.configReload = { hotReloadStatus: "active" };

  const report = await buildStatusCommandReportData({
    ...params,
    opts: { deep: true },
    health,
  });

  expect(report.healthRows?.some(({ Item }) => Item === "Context engine")).toBe(false);
  expect(report.healthRows?.some(({ Item }) => Item === "Config hot reload")).toBe(false);
});
```

Why this is the right current-main boundary:
- it exercises `buildStatusCommandReportData`, not only the formatter helpers;
- it preserves current main's Gateway / SQLite WAL / event-loop rows;
- it distinguishes operational WARN from intentionally disabled channel OFF;
- it does not reintroduce the already-moved formatter ownership.

Suggested focused validation:

```bash
pnpm test src/commands/status.command-report-data.test.ts --run
pnpm test src/commands/health-format.test.ts --run
```

Status: source-reviewed against current main; not executed in this environment.
