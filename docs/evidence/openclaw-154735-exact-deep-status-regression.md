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

Do **not** feed these operational lines through the generic channel-row parser.

Current main classifies any detail beginning with `disabled` as `OFF`, so
`Config hot reload: disabled (watcher retries exhausted...)` would be misclassified even
though watcher exhaustion is an operational failure.

Keep channel parsing unchanged, then append these formatter outputs explicitly as `WARN` rows:

```ts
for (const line of [
  formatContextEngineHealthLine(params.health),
  formatDeliveryQueueHealthLine(params.health),
  formatConfigReloadHealthLine(params.health),
]) {
  if (!line) {
    continue;
  }
  const colon = line.indexOf(":");
  rows.push({
    Item: line.slice(0, colon),
    Status: theme.warn("WARN"),
    Detail: line.slice(colon + 1).trim(),
  });
}
```

This matches the reviewed branch's semantic distinction:

- intentionally disabled channel -> `OFF`;
- exhausted config-reload watcher -> `WARN`;
- quarantined context engine -> `WARN`;
- delivery dead-letter / ingress pressure -> `WARN`.

The exact regression above is useful precisely because it catches the accidental
`disabled => OFF` misclassification.

Suggested focused validation:

```bash
node scripts/run-vitest.mjs src/commands/status.command-report-data.test.ts --run
node scripts/run-vitest.mjs src/commands/health-format.test.ts --run
```

Self-check notes (2026-10-07):
- current test file already imports `expectDefined`, `stripAnsi`, `buildStatusCommandReportData`, and `createStatusCommandReportDataParams`;
- current protocol schema confirms the context-engine and config-reload fixture shapes above;
- using object spread avoids mutating a possibly narrowed fixture object;
- expected row details exactly match the current shared formatter output after explicit operational-row splitting at the first colon;
- source-reviewed, not executed.

Status: source-reviewed against current main; not executed in this environment.


## Correction note

An earlier version of this artifact proposed appending the operational formatter strings into the
generic `healthLines` parser. That was incorrect for config reload: current main maps details
starting with `disabled` to `OFF`, while an exhausted reload watcher is an operational failure
and must be `WARN`.

The corrected integration above mirrors the reviewed branch's explicit WARN-row handling.
