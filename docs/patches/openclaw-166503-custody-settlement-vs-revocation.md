# OpenClaw #166503 — distinguish native custody settlement from discovery revocation

Target head: `efedc3db2f14d5e5cb35cff008a13433f14fecd9`

## Root cause

Fresh ClawSweeper Revision 2 reports exact-head Gateway failures at the post-cleanup check in
`withSessionHistoryWorkerReadCandidates()`.

Current candidate:

```ts
if (
  revoked ||
  [...retained.values()].some(
    (resource) =>
      resource.revoked ||
      Boolean(resource.closing) ||
      !resource.nativeSequences.has(lane),
  )
) {
  throw new WorkerTaskError(
    "Session reader custody was revoked during discovery cleanup",
    "unavailable",
  );
}
```

The last predicate conflates two different facts:

```text
resource.nativeSequences.has(lane)
= this worker generation still carries native reader custody

resource.revoked / top-level revoked
= discovery authority is no longer current
```

Native custody can disappear normally.

Existing production owners explicitly clear it on success:

- `clearClosedDatabaseCustody()` removes a lane sequence after a successful worker reply reports a
  closed history database;
- `releaseRetiredDatabaseCustody()` removes lane sequences when a worker generation retires;
- neither operation sets `resource.revoked`.

Therefore:

```text
missing native sequence
!=
revoked discovery authority
```

## Minimal production repair

Keep the real authority checks; remove the custody-presence requirement from the post-cleanup
revocation predicate.

`src/config/sessions/session-transcript-worker-resources.ts`

```diff
         if (
           revoked ||
           [...retained.values()].some(
-            (resource) =>
-              resource.revoked || Boolean(resource.closing) || !resource.nativeSequences.has(lane),
+            (resource) => resource.revoked || Boolean(resource.closing),
           )
         ) {
```

Do **not** remove the earlier native-custody filter used to build `retained`:

```ts
if (resource.revoked || resource.closing || !resource.nativeSequences.has(lane)) {
  continue;
}
```

That earlier condition answers a different question:

> Which database resources are still physically retained by this worker before `closeResources()`?

The post-cleanup check answers:

> Did authority become invalid while cleanup awaited?

Those must not share the same predicate.

## Why authority remains protected

During the entire operation, `registerOpenClawAgentDatabaseReadCandidateResource()` registrations
remain active until the final release path.

Their `revoke` callback sets the enclosing:

```ts
revoked = true;
```

and `assertCurrent()` refuses when that flag is set.

For resources that were retained before cleanup:

- lexical aliases were already validated with `isSessionStoreReadCandidateCurrent()`;
- aliases are retained on the known resource before native cleanup;
- `resource.revoked` still captures resource-level invalidation;
- `resource.closing` still rejects an externally closing resource;
- after cleanup, the enclosing operation calls `assertCurrent()` again before returning.

So dropping only `!resource.nativeSequences.has(lane)` does not weaken path/revocation authority.
It only stops treating normal native-reader settlement as revocation.

## Regression

The exact-head CI failures already exercise the production composition:

- `src/gateway/session-create-display-name.test.ts`
- Gateway RPC cohorts from jobs `112676315430`, `112676315452`, and `112676315265`

The repair should first be validated by rerunning those original failing cohorts, not only a new
standalone unit test.

Also retain focused custody tests:

```bash
pnpm test   src/config/sessions/session-transcript-worker-lanes.test.ts   src/config/sessions/session-transcript-search.worker.test.ts   --maxWorkers=1

pnpm test src/gateway/session-create-display-name.test.ts --maxWorkers=1
```

### Focused invariant worth adding

A small owner-level regression can make the distinction durable:

```text
retained resource has native custody
→ normal successful close/retirement clears nativeSequences
→ resource.revoked remains false
→ candidate registration remains current
→ discovery result is returned

same setup, but candidate/resource revoke fires while cleanup awaits
→ discovery still rejects unavailable
```

This prevents a future optimization from re-coupling custody presence to authority.

## Non-goals

- do not remove `resource.revoked`;
- do not remove `resource.closing`;
- do not weaken candidate path / alias identity checks;
- do not change worker retirement policy;
- do not introduce a persisted field or migration.

Status: source-reviewed against `efedc3db2f14d5e5cb35cff008a13433f14fecd9`; not executed by
`hippoley`.
