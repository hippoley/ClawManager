# OpenClaw #166478 — preserve Linux worker termination while keeping Darwin census-free

Target head: `73440925ad20d763fd69b862418b8836e0ff3228`

## Root cause

The candidate changes every non-Windows worker shutdown to:

```ts
process.kill(
  isOwnedProcessGroupGone(process.pid) ? process.pid : -process.pid,
  "SIGKILL",
);
```

That is valid only when the negative-PID probe is authoritative for group ownership.

On supported Linux hosts that deny group-directed kill syscalls, the existing native fixture proves:

```ts
process.kill(-2147483647, 0) -> EPERM
```

and `isOwnedProcessGroupGone()` deliberately treats EPERM as "not known gone", not as proof that
the current worker leads that group.

The candidate therefore selects `-process.pid`, and the same kernel policy denies the SIGKILL.

## Smallest production repair

Keep the new census-free path only on Darwin, where it replaces the expensive `ps` ownership
census that motivated this PR.

Preserve the existing worker termination path everywhere else.

`src/worker/worker-process.ts`

```diff
     terminateOwnedTree: () => {
-      if (process.platform === "win32") {
+      if (process.platform !== "darwin") {
         signalProcessTree(process.pid, "SIGKILL");
         return;
       }
       // Anchored applications share their owner's group; direct workers may lead their own.
       // Exec relays start parent-loss cleanup only after this process dies, so decide by
-      // syscall: a ps census here can stall past their cleanup budget.
+      // syscall on Darwin: a ps census here can stall past their cleanup budget.
       process.kill(isOwnedProcessGroupGone(process.pid) ? process.pid : -process.pid, "SIGKILL");
     },
```

No other production file needs to change.

## Why this preserves Linux correctly

The existing `signalProcessTree(process.pid, "SIGKILL")` already avoids the Darwin latency problem
on Linux:

```text
signalProcessTree
  -> isProcessGroupLeader(pid)
  -> Linux reads /proc/<pid>/stat
  -> no synchronous ps census on the common path
```

For an attached Linux worker:

```text
/proc says pgid != pid
-> useGroupKill = false
-> signal positive worker PID
```

So the kernel policy that denies negative/group-directed kill does not block the worker's own
immediate termination.

For a genuinely detached/direct worker that is its own group leader, the previous behavior remains
unchanged.

Windows also remains on its pre-PR `signalProcessTree` path.

## Regression owner

The strongest regression should combine the worker supervisor-loss path with the repository's
existing real Linux group-signal-denial kernel contract.

Existing proof owner:

`src/process/supervisor/service-child-subreaper.test-support.ts`

That fixture already installs a real seccomp filter that:

- denies group-directed `kill` operations with EPERM;
- verifies `process.kill(0, 0)` is denied;
- verifies `process.kill(-2147483647, 0)` is denied;
- keeps supported direct/native cleanup working.

### Required behavior

Under that same filter:

```text
launch an attached worker with real supervisor lifetime
start the existing background heartbeat/exec fixture
end the worker anchor/supervisor lifetime
=> worker process exits
=> background relay cleanup settles
=> worker capacity is released
```

The important assertion is not merely that group probing returns EPERM. It must prove the worker
shutdown path still reaches positive-PID termination under the real denial policy.

## Focused source-level guard

If a small deterministic test seam is desirable in addition to the real-kernel fixture, extract only
the platform choice, not process ownership policy:

```ts
function terminateWorkerOwnedTree(): void {
  if (process.platform !== "darwin") {
    signalProcessTree(process.pid, "SIGKILL");
    return;
  }
  process.kill(isOwnedProcessGroupGone(process.pid) ? process.pid : -process.pid, "SIGKILL");
}
```

Then the ordinary worker runtime test can assert the Darwin-specific path is the only new special
case while Linux retains the existing owner.

Do not add a new Linux ownership algorithm.

## Data-model compatibility note

The review's data-model detector reports:

`serialized state: src/worker/worker-runtime-background-exec.suite.ts`

but that file is test support / behavior fixture, not a persisted-state writer.

The production delta in this repair:

- changes no config key;
- changes no SQLite schema/table/column;
- changes no serialized runtime record;
- changes no protocol message;
- adds no migration/backfill/repair owner.

The appropriate compatibility statement is:

```text
No stored-data contract change; no migration or backfill required.
```

## Acceptance

Use the reviewer's existing acceptance commands:

```bash
node scripts/run-vitest.mjs src/worker/worker.runtime.test.ts
node scripts/run-vitest.mjs src/process/supervisor/service-child-subreaper.real.test.ts
node scripts/check-changed.mjs
git diff --check
```

For the native Linux acceptance, record that the group-signal-denial fixture is active and that
supervisor loss still terminates the attached worker / releases capacity.

Status: source-reviewed against `73440925ad20d763fd69b862418b8836e0ff3228`; not executed.
