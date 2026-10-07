# OpenClaw #163138 — preserve expiry wake across failed commit overlap

Target head: `fea95c9d35d6c7d605e321f51a98e74a71a77ce6`

## Fresh blocker

ClawSweeper's latest review found a real lost-wake race in the scheduled-expiry repair.

Current candidate:

```ts
private expireScheduled(id: string): Promise<void> | void {
  const entry = this.entries.get(id);
  if (!entry || entry.record.status !== "pending" || entry.committing) {
    return this.drain();
  }
  const remainingMs = entry.expiresAtMonotonicMs - this.scheduler.monotonicNow();
  if (remainingMs > 0) {
    entry.job = this.scheduler.schedule(...);
    return this.drain();
  }
  this.expire(id);
  return this.drain();
}
```

If an early wall-clock wake fires while `resolveWithCommit()` is awaiting a write:

1. GatewayScheduler has already removed the one-shot job before invoking the callback.
2. `entry.committing` causes `expireScheduled()` to return immediately.
3. No replacement wake is installed.
4. The commit rejects **before** the monotonic deadline.
5. `resolveWithCommit().finally` clears `committing` and calls `get(id)`.
6. Because the deadline is still in the future, `get()` leaves the question pending.
7. No expiry job remains, so a timeout-free waiter can remain pending past the deadline until another
   read forces expiry.

## Minimal production repair

Re-arm any early wake **before** treating an active commit as non-expirable.

`src/gateway/question-manager.ts`

```diff
   private expireScheduled(id: string): Promise<void> | void {
     const entry = this.entries.get(id);
-    if (!entry || entry.record.status !== "pending" || entry.committing) {
+    if (!entry || entry.record.status !== "pending") {
       return this.drain();
     }
     const remainingMs = entry.expiresAtMonotonicMs - this.scheduler.monotonicNow();
     if (remainingMs > 0) {
       // Wall clock jumped forward; the monotonic deadline has not passed.
       // Re-arm the scheduled wake for the remaining monotonic duration.
       entry.job = this.scheduler.schedule({
         id: `${this.scheduleId}:${id}`,
         delayMs: remainingMs,
         run: () => this.expireScheduled(id),
       });
       return this.drain();
     }
+    if (entry.committing) {
+      return this.drain();
+    }
     this.expire(id);
     return this.drain();
   }
```

This preserves the intended ownership rule:

- before the monotonic deadline: always retain an automatic future wake;
- at/after the deadline while no commit owns the mutation: expire normally;
- at/after the deadline while a commit is still active: do not race the commit; its existing
  `assertCurrent()` / `finally -> get(id)` path resolves the terminal state.

No new state field or scheduler API is needed.

## Exact regression

Use the existing fake scheduler clock and `createDeferredCore()`.

The test must **not** call `get()` or `list()` after the commit fails, because those reads can
force expiry and hide the lost-wake bug.

Add to `src/gateway/question-manager.test.ts`:

```ts
it("rearms an early expiry wake while a commit is running and expires after the commit fails", async () => {
  const commitGate = createDeferredCore();
  const commitEntered = createDeferredCore();

  const record = manager.request({
    questions,
    timeoutMs: 1_000,
  });
  const waiting = manager.waitAnswer(record.id);

  const resolving = manager.resolveWithCommit(record.id, answers, undefined, {
    commit: async () => {
      commitEntered.resolve();
      await commitGate.promise;
      throw new Error("synthetic commit failure");
    },
  });

  await commitEntered.promise;

  // Simulate the scheduler being woken early only because the wall clock jumped.
  // Monotonic time remains before the question deadline.
  await clock.wakeDueToWallClock();

  // Observation only: do not force expiry through get()/list().
  expect(manager.observe(record.id)?.record.status).toBe("pending");
  expect(clock.armedAtMs).not.toBeNull();

  // Fail the write while the monotonic deadline is still in the future.
  commitGate.resolve();
  await expect(resolving).rejects.toThrow("synthetic commit failure");

  // The failed commit's finally may inspect the question, but should leave it pending.
  // Most importantly, the early wake must already have installed a replacement job.
  expect(manager.observe(record.id)?.record.status).toBe("pending");
  expect(clock.armedAtMs).not.toBeNull();

  let settled = false;
  void waiting.then(() => {
    settled = true;
  });

  await clock.advanceBy(999);
  await Promise.resolve();
  expect(settled).toBe(false);
  expect(manager.observe(record.id)?.record.status).toBe("pending");

  await clock.advanceBy(1);
  await expect(waiting).resolves.toEqual({ status: "expired" });
  expect(manager.observe(record.id)?.record.status).toBe("expired");
});
```

### Why this catches the current bug

On `fea95c9d`:

```text
wakeDueToWallClock()
-> expireScheduled()
-> entry.committing === true
-> return without schedule()
-> clock.armedAtMs === null
```

The regression fails immediately on that assertion.

With the repair:

```text
early wake
-> remainingMs > 0
-> replacement job installed
-> commit fails before deadline
-> question remains pending
-> replacement job fires at monotonic deadline
-> timeout-free waiter receives expired
```

No later `get()` or `list()` is needed to create the terminal fact.

## Acceptance

Use the reviewer's current acceptance set:

```bash
pnpm test src/gateway/question-manager.test.ts --maxWorkers=1
pnpm test src/agents/tools/ask-user-tool.test.ts --maxWorkers=1
git diff --check
```

Status: source-reviewed against `fea95c9d35d6c7d605e321f51a98e74a71a77ce6`; not executed.
