# OpenClaw #166425 — interrupted recovery-seal classifier + consumer repair

Target head: `19d2db119dab26136d8264e8608e33e2416dbda6`

## Problem

The new fallback publisher deliberately has a two-name window:

```text
manifest.json.partial
        |
        | publish link-or-copy
        v
manifest.json
        |
        | verify source identity
        | unlink partial
        v
sealed capture
```

If the process stops after target publication but before source removal, both names remain.

This is not a sealed generation. A successful publication protocol always removes
`manifest.json.partial`.

Current consumers disagree with that protocol:

- `readUpdateRecoveryBackups()` sees `manifest.json`, enters strict
  `hardlinks: "reject"` reading, and can throw for the interrupted hard-link form,
  hiding otherwise healthy recovery sets.
- `retireExpiredStandaloneDoctorCaptures()` reads `manifest.json` directly and can
  later delete the directory even though publication cleanup never completed.

## Design rule

Treat the exact two-name state as **incomplete evidence**, not as a sealed manifest.

Do not weaken the strict manifest reader.

In particular:

```text
manifest.json only, ordinary file
  => normal strict sealed validation

manifest.json only, unrelated hardlink
  => existing strict rejection remains

manifest.json.partial only
  => existing incomplete classification remains

manifest.json + manifest.json.partial
  => incomplete / never auto-retire
```

This rule also covers `exclusive-copy` fallback, not only hard-link fallback. The consumer does
not need to infer which filesystem publication method was used. The presence of the temporary
source name is itself proof that the capture-owned seal protocol did not finish.

## Shared classifier

Put the protocol predicate beside the publisher that owns the temporary/final naming contract.

`src/infra/update-recovery-capture-publication.ts`

```diff
 import type { BigIntStats } from "node:fs";
 import fs from "node:fs/promises";
 import path from "node:path";
@@
 import {
   publishFileExclusive,
   requireDirectorySync,
   syncDirectory,
 } from "./directory-durability.js";
+import { hasErrnoCode } from "./errno.js";
+
+export const UPDATE_RECOVERY_MANIFEST_NAME = "manifest.json";
+export const UPDATE_RECOVERY_PARTIAL_MANIFEST_NAME = "manifest.json.partial";
+
+async function lstatOrMissing(pathname: string) {
+  try {
+    return await fs.lstat(pathname);
+  } catch (error) {
+    if (hasErrnoCode(error, "ENOENT")) {
+      return undefined;
+    }
+    throw error;
+  }
+}
+
+/**
+ * A final manifest with its capture-owned temporary source still present is
+ * retained incomplete evidence. Successful publication always removes the
+ * temporary source after the exclusive target is durable.
+ *
+ * This predicate does not bless either file as readable/restorable.
+ */
+export async function hasInterruptedUpdateRecoveryManifestSeal(
+  directory: string,
+): Promise<boolean> {
+  const finalPath = path.join(directory, UPDATE_RECOVERY_MANIFEST_NAME);
+  const partialPath = path.join(directory, UPDATE_RECOVERY_PARTIAL_MANIFEST_NAME);
+  const [finalEntry, partialEntry] = await Promise.all([
+    lstatOrMissing(finalPath),
+    lstatOrMissing(partialPath),
+  ]);
+  if (!finalEntry || !partialEntry) {
+    return false;
+  }
+  // Do not downgrade malformed/symlink publication into a benign incomplete
+  // state. Consumers retain their existing fail-closed behavior for those.
+  if (!finalEntry.isFile() || !partialEntry.isFile()) {
+    throw new Error(
+      `Interrupted update recovery manifest seal has an unsupported file kind: ${directory}`,
+    );
+  }
+  return true;
+}
```

The existing publisher should use the constants too:

```diff
- const manifestPath = path.join(directory, "manifest.json");
- const temporaryManifestPath = path.join(directory, "manifest.json.partial");
+ const manifestPath = path.join(directory, UPDATE_RECOVERY_MANIFEST_NAME);
+ const temporaryManifestPath = path.join(
+   directory,
+   UPDATE_RECOVERY_PARTIAL_MANIFEST_NAME,
+ );
```

Import the constants where appropriate rather than creating a second spelling owner.

## Consumer 1 — recovery discovery

`src/infra/update-recovery-backup-reader.ts`

Before constructing the strict `safeRoot(..., hardlinks: "reject")` reader:

```diff
+import {
+  hasInterruptedUpdateRecoveryManifestSeal,
+  UPDATE_RECOVERY_MANIFEST_NAME,
+} from "./update-recovery-capture-publication.js";
@@
- const manifestPath = path.join(directory, "manifest.json");
+ const manifestPath = path.join(directory, UPDATE_RECOVERY_MANIFEST_NAME);
  const manifestEntry = entry?.isDirectory() ? await statOrMissing(manifestPath) : undefined;
@@
  if (entry?.isDirectory() && !manifestEntry && /^[a-zA-Z0-9_-]{1,128}$/u.test(captureId)) {
    result.push({ kind: "incomplete", directory });
    continue;
  }
+ if (
+   entry?.isDirectory() &&
+   manifestEntry?.isFile() &&
+   (await hasInterruptedUpdateRecoveryManifestSeal(directory))
+ ) {
+   result.push({ kind: "incomplete", directory });
+   continue;
+ }
  if (!entry?.isDirectory() || !manifestEntry?.isFile()) {
    throw new Error(...);
  }
  const source = await safeRoot(directory, { symlinks: "reject", hardlinks: "reject" });
```

Important: the classifier runs **before** strict manifest reading, and only changes the exact
two-name protocol state. A hard-linked `manifest.json` without the capture-owned
`.partial` sibling still reaches the existing hardlink rejection.

## Consumer 2 — standalone Doctor retirement

`src/infra/update-recovery-baseline-capture.ts`

Before reading/parsing a final manifest:

```diff
 import {
+  hasInterruptedUpdateRecoveryManifestSeal,
   publishUpdateRecoveryCaptureFile,
+  UPDATE_RECOVERY_MANIFEST_NAME,
 } from "./update-recovery-capture-publication.js";
@@
     const directory = path.join(root, name);
     try {
       const entry = await fs.lstat(directory);
       if (!entry.isDirectory() || entry.isSymbolicLink()) {
         continue;
       }
+      if (await hasInterruptedUpdateRecoveryManifestSeal(directory)) {
+        continue;
+      }
       let raw: string;
       try {
-        raw = await fs.readFile(path.join(directory, "manifest.json"), "utf8");
+        raw = await fs.readFile(
+          path.join(directory, UPDATE_RECOVERY_MANIFEST_NAME),
+          "utf8",
+        );
       } catch (error) {
```

This preserves interrupted evidence regardless of age. It also makes the retirement contract
match its own comment:

> Only sealed, unassociated standalone originals are eligible for retirement.

If the classifier throws because the sibling is a symlink/directory/non-regular object, the
existing retirement `catch` converts that into a warning and preserves the directory, which is
the correct fail-closed behavior.

## Regression 1 — healthy status survives interrupted hard-link seal

Extend `src/cli/update-cli/status.recovery.test.ts`.

Use the existing `capture()` helper:

```ts
it("keeps healthy recovery sets visible beside an interrupted fallback seal", async () => {
  const healthy = await capture("11111111-1111-4111-8111-111111111111");
  await terminal(healthy, "committed");

  const interrupted = await capture("22222222-2222-4222-8222-222222222222");
  const partialPath = path.join(interrupted.directory, "manifest.json.partial");

  // Recreate the exact hard-link fallback interruption window:
  // partial is the publication source, final is the linked target.
  await fs.rename(interrupted.manifestPath, partialPath);
  await fs.link(partialPath, interrupted.manifestPath);

  expect((await fs.lstat(partialPath)).ino).toBe(
    (await fs.lstat(interrupted.manifestPath)).ino,
  );

  await updateStatusCommand({ json: true });

  expect(result()).not.toHaveProperty("recoverySetsError");
  expect(result().recoverySets).toEqual(
    expect.arrayContaining([
      expect.objectContaining({
        runId: healthy.manifest.runId,
        manifestPath: healthy.manifestPath,
        status: "stale",
      }),
      expect.objectContaining({
        directory: interrupted.directory,
        status: "incomplete",
      }),
    ]),
  );

  // Discovery must be read-only.
  expect(await fs.readFile(partialPath, "utf8")).toBe(interrupted.raw);
  expect(await fs.readFile(interrupted.manifestPath, "utf8")).toBe(interrupted.raw);
});
```

Keep the existing `manifest-hardlink` fault test unchanged. It creates a hard-linked final
without `manifest.json.partial` and must still produce `recoverySetsError`.

Also add an exclusive-copy form:

```ts
it("classifies a copied final with retained partial as incomplete", async () => {
  const interrupted = await capture();
  const partialPath = path.join(interrupted.directory, "manifest.json.partial");
  await fs.rename(interrupted.manifestPath, partialPath);
  await fs.copyFile(partialPath, interrupted.manifestPath);

  await updateStatusCommand({ json: true });

  expect(result()).not.toHaveProperty("recoverySetsError");
  expect(result().recoverySets).toEqual([
    expect.objectContaining({
      directory: interrupted.directory,
      status: "incomplete",
    }),
  ]);
});
```

## Regression 2 — retirement never removes interrupted seal

Extend the existing test:

`retires only expired sealed standalone Doctor captures and preserves incomplete or linked copies`

Add:

```ts
const interruptedId = `doctor-${randomUUID()}`;
const interrupted = await sealed(interruptedId, 31);
const interruptedPartial = path.join(interrupted, "manifest.json.partial");

// Hard-link form of the publication interruption.
await fs.link(path.join(interrupted, "manifest.json"), interruptedPartial);
```

After retirement:

```ts
expect(result.retired).toEqual([expired]);
expect((await fs.lstat(interrupted)).isDirectory()).toBe(true);
expect(await fs.readFile(interruptedPartial)).toEqual(
  await fs.readFile(path.join(interrupted, "manifest.json")),
);
```

Optionally parameterize this case over `hardlink` and `copy` to cover both fallback methods.

## Why not inspect nlink / inode as the primary policy?

The consumer contract is stronger and simpler than the filesystem implementation detail:

```text
successful capture seal
=> temporary manifest name is gone
```

Using only `nlink === 2` would miss the exclusive-copy fallback, even though it has the same
interruption window. Requiring source/final identity would therefore fix only part of the new
reachable state.

The strict reader remains responsible for rejecting unrelated hard links once no temporary
publication source exists.

## Focused validation

```bash
node scripts/run-vitest.mjs src/infra/update-recovery-capture-publication.test.ts --run
node scripts/run-vitest.mjs src/infra/update-database-backup.test.ts --run
node scripts/run-vitest.mjs src/cli/update-cli/status.recovery.test.ts --run
pnpm tsgo:core
git diff --check
```

Status: source-reviewed against head
`19d2db119dab26136d8264e8608e33e2416dbda6`; not executed.
