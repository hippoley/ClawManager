# OpenClaw #154735 — current-main integration patch

Current main already owns these shared formatters in `src/commands/health-format.ts`:

- `formatContextEngineHealthLine`
- `formatConfigReloadHealthLine`
- `formatDeliveryQueueHealthLine`

The remaining omission is in `buildStatusHealthRows()`: it appends delivery-queue health, but not the context-engine or exhausted-reload operational warnings.

## Minimal current-main patch

```diff
diff --git a/src/commands/status.command-sections.ts b/src/commands/status.command-sections.ts
--- a/src/commands/status.command-sections.ts
+++ b/src/commands/status.command-sections.ts
@@
-import { formatDeliveryQueueHealthLine, formatHealthChannelLines } from "./health-format.js";
+import {
+  formatConfigReloadHealthLine,
+  formatContextEngineHealthLine,
+  formatDeliveryQueueHealthLine,
+  formatHealthChannelLines,
+} from "./health-format.js";
@@
   const healthLines = formatHealthChannelLines(params.health, { accountMode: "all" });
-  const deliveryQueueLine = formatDeliveryQueueHealthLine(params.health);
-  if (deliveryQueueLine) {
-    healthLines.push(deliveryQueueLine);
-  }
   for (const line of healthLines) {
@@
     rows.push({ Item: item, Status: status, Detail: detail });
   }
+
+  // Operational failures are WARN rows, unlike intentionally disabled channels (OFF).
+  for (const line of [
+    formatContextEngineHealthLine(params.health),
+    formatDeliveryQueueHealthLine(params.health),
+    formatConfigReloadHealthLine(params.health),
+  ]) {
+    if (!line) {
+      continue;
+    }
+    const colon = line.indexOf(":");
+    if (colon === -1) {
+      continue;
+    }
+    rows.push({
+      Item: line.slice(0, colon).trim(),
+      Status: theme.warn("WARN"),
+      Detail: line.slice(colon + 1).trim(),
+    });
+  }
   return rows;
 }
```

## Focused regression

Add to `src/commands/status.command-report-data.test.ts`:

```ts
it("surfaces auxiliary operational degradation in deep status", async () => {
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

  expect(
    report.healthRows?.map(({ Item, Status, Detail }) => ({
      Item,
      Status: stripAnsi(Status),
      Detail,
    })),
  ).toEqual(
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

Keep the existing SQLite WAL and event-loop rows untouched. The point of this refresh is to integrate with current main's newer status-row owners, not to reintroduce the older formatter move.

## Suggested focused validation

```bash
pnpm test src/commands/health-format.test.ts
pnpm test src/commands/status.command-report-data.test.ts
pnpm test src/commands/health.test.ts
```

## Scope

No schema/config/state-write change.
No channel classification change.
No queue-warning semantic change.
No runtime recovery behavior change.

Status: source-reviewed against current main on 2026-10-07; not executed in this environment.
