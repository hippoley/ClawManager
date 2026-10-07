# OpenClaw #154735 — exact current-main deep-status regression

Target: current `main` after reusing the shared formatter owners that have already landed.

Add this focused regression to `src/commands/status.command-report-data.test.ts`:

```ts
it("shows context-engine quarantine and exhausted config reload in deep status", async () => {
  const params = createStatusCommandReportDataParams();
  const health = {
    ...expectDefined(params.health, "health fixture"),
    contextEngines: {
      quarantined: [
        {
          engineId: "lossless-claw",
          owner: "plugin:lossless-claw",
          operation: "assemble",
          reason: "database corrupt",
          failedAt: Date.now(),
        },
      ],
    },
    configReload: { hotReloadStatus: "disabled" as const },
  };

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
  const health = {
    ...expectDefined(params.health, "health fixture"),
    contextEngines: { quarantined: [] },
    configReload: { hotReloadStatus: "active" as const },
  };

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
- it uses the actual protocol-backed `HealthSummary` shapes:
  - `contextEngines.quarantined[] = { engineId, owner?, operation, reason, failedAt }`
  - `configReload.hotReloadStatus = "active" | "disabled"`
- it preserves current main's Gateway / SQLite WAL / event-loop rows;
- it distinguishes operational WARN from intentionally disabled channel OFF;
- it does not reintroduce the already-moved formatter ownership.

The production integration should import and append the existing shared formatter owners:

```ts
import {
  formatConfigReloadHealthLine,
  formatContextEngineHealthLine,
  formatDeliveryQueueHealthLine,
  formatHealthChannelLines,
} from "./health-format.js";
```

Then, alongside the existing delivery-queue line:

```ts
for (const operationalLine of [
  formatContextEngineHealthLine(params.health),
  formatDeliveryQueueHealthLine(params.health),
  formatConfigReloadHealthLine(params.health),
]) {
  if (operationalLine) {
    healthLines.push(operationalLine);
  }
}
```

The existing row parser will classify both new warning strings as `WARN` because neither begins with an OK/OFF/LINKED prefix.

Suggested focused validation:

```bash
node scripts/run-vitest.mjs src/commands/status.command-report-data.test.ts --run
node scripts/run-vitest.mjs src/commands/health-format.test.ts --run
```

Self-check notes (2026-10-07):
- current test file already imports `expectDefined`, `stripAnsi`, `buildStatusCommandReportData`, and `createStatusCommandReportDataParams`;
- current protocol schema confirms the context-engine and config-reload fixture shapes above;
- using object spread avoids mutating a possibly narrowed fixture object;
- expected row details exactly match the current shared formatter output after the table parser strips the prefix before the first colon;
- source-reviewed, not executed.

Status: source-reviewed against current main; not executed in this environment.
