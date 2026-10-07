# OpenClaw #166487 — align stale incognito append tests with prefix consistency

Target head: `79313e24dc13910138b84fe3a0ce541faed5befe`

Fresh ClawSweeper Revision 4 resolves the earlier replay P1 and leaves one bounded test-only repair:
two incognito-history cases still expect append invalidation even though ordinary detached context reads now
intentionally preserve the captured prefix.

The correct contract is:

```text
ordinary append during async consumption
→ consumer resolves with the original captured snapshot
→ appended content is not injected into that snapshot

revocation / abort / released execution reference
→ still reject

prepared replay admission
→ still exact, not prefix-tolerant
```

## 1. history-wiring sibling

`src/state/openclaw-agent-execution-incognito.history-wiring.test-support.ts`

Current stale branch:

```ts
mode === "write"
  ? expect(work).rejects.toBeInstanceOf(SessionTranscriptReadFenceError)
```

Replace the success expectation for both `unchanged` and `write` with a snapshot assertion.

One low-churn form:

```diff
-      const settled =
-        mode === "unchanged"
+      const settled =
+        mode === "unchanged" || mode === "write"
           ? expect(work).resolves.toEqual(
               expect.arrayContaining([
                 expect.objectContaining({
                   message: expect.objectContaining({
                     content: [{ type: "text", text: "private context before consumer" }],
                   }),
                 }),
               ]),
             )
-          : mode === "write"
-            ? expect(work).rejects.toBeInstanceOf(SessionTranscriptReadFenceError)
-            : expect(work).rejects.toThrow(
+          : expect(work).rejects.toThrow(
               mode === "release"
                 ? "reference is released"
                 : mode === "abort"
                   ? "context admission revoked"
                   : "context grant revoked",
             );
```

Then keep the existing append at:

```ts
await append(session, "context changed while consumer awaited");
```

After settlement, make the preserved-prefix exclusion explicit:

```ts
if (mode === "write") {
  const resolved = await work;
  expect(JSON.stringify(resolved)).toContain("private context before consumer");
  expect(JSON.stringify(resolved)).not.toContain("context changed while consumer awaited");
}
```

If awaiting `work` a second time is considered noisy, store the promise result once before the expectation;
the important part is that the write case proves both inclusion of the captured prefix and exclusion of the later append.

Do not change:

- revoke behavior;
- abort behavior;
- release/join behavior;
- single-consumer invocation count.

## 2. Codex history sibling

`src/state/openclaw-agent-execution-incognito.history.test.ts`

Current stale expectation:

```ts
await expect(
  reader.nativeContext(scope, async (messages) => {
    const result = [...messages];
    await append(target, "invalidates snapshot");
    return result;
  }),
).rejects.toBeInstanceOf(SessionTranscriptReadFenceError);
```

Change it to assert that the original snapshot resolves and excludes the later append:

```ts
const snapshot = await reader.nativeContext(scope, async (messages) => {
  const result = [...messages];
  await append(target, "later appended content");
  return result;
});

expect(snapshot).toMatchObject([
  {
    content: [{ type: "text", text: "snapshot content" }],
  },
]);
expect(JSON.stringify(snapshot)).not.toContain("later appended content");
```

Use the repository's exact message shape if this fixture includes role/timestamp fields;
the invariant is the same: the returned snapshot is the one captured before the append.

The second half of the same test must stay unchanged:

```text
borrowed execution reference released during async consumer
→ consumer rejects
→ release waits for joined lifetime
```

That verifies prefix tolerance does not weaken execution-reference authority.

## Why this is the right final repair

Latest production owner:

`src/config/sessions/session-transcript-anchor-read.kernel.ts`

already selects:

```ts
replay ? "exact" : "prefix"
```

So these two tests should not try to restore exactness globally.

The fresh review explicitly confirms:

- prepared replay strictness is restored;
- the two CI failures are stale sibling expectations;
- the intended repair is test alignment, not another production check.

## Acceptance

```bash
pnpm test src/state/openclaw-agent-execution-incognito.history.test.ts --maxWorkers=1
pnpm test src/agents/sessions/session-manager-model-context.test.ts --maxWorkers=1
pnpm test src/agents/embedded-agent-runner/run/attempt-session-replay.test.ts --maxWorkers=1
git diff --check
```

Status: source-reviewed against `79313e24dc13910138b84fe3a0ce541faed5befe`; not executed by `hippoley`.
