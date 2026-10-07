# OpenClaw #166514 — match command listings after recall normalization

Target head: `846873fde872558e8e8946a59977b5203ecd487c`

## Root cause

`recordShortTermRecalls()` normalizes every snippet before storage:

```ts
const rawSnippet = normalizeSnippet(result.snippet);
```

and `normalizeSnippet()` is:

```ts
raw.trim().replace(/\s+/g, " ")
```

Therefore a real porcelain listing such as:

```text
 M tracked.md
?? new.md
```

reaches dreaming as:

```text
M tracked.md ?? new.md
```

The current predicate still expects newline boundaries:

```ts
/(?:^|\\n|\n)\?\?\s+\S/
```

so the production-normalized form is admitted.

The current regression bypasses this boundary by writing raw recall rows and by making the fixture
also match the independent `cron-self-heal` exclusion.

## Smallest production repair

Match the normalized token stream rather than reconstructing line boundaries.

A bounded helper avoids treating arbitrary question marks or prose as command output:

```ts
const GIT_PORCELAIN_STATUS_TOKEN_RE = /^(?:\?\?|!!|[MADRCU]{1,2})$/;

function isNormalizedGitStatusListing(raw: string): boolean {
  const snippet = raw.replace(/^(?:User|Assistant):\s*/i, "").trim();
  const tokens = snippet.split(/\s+/);
  let pairs = 0;
  let firstMarker: string | undefined;

  for (let index = 0; index + 1 < tokens.length; index += 1) {
    const marker = tokens[index]!;
    if (!GIT_PORCELAIN_STATUS_TOKEN_RE.test(marker)) {
      continue;
    }
    const candidatePath = tokens[index + 1]!;
    if (GIT_PORCELAIN_STATUS_TOKEN_RE.test(candidatePath)) {
      continue;
    }
    firstMarker ??= marker;
    pairs += 1;
    index += 1;
  }

  if (pairs >= 2) {
    return true;
  }

  // A single untracked/ignored entry is already an unambiguous status record.
  return pairs === 1 && (firstMarker === "??" || firstMarker === "!!");
}
```

Then the shared predicate becomes:

```ts
return (
  /(?:^|\/)\d{4}-\d{2}-\d{2}-reconcile-report\.md$/i.test(sourcePath) ||
  /^(?:Broken Links?|Orphan Notes?|Duplicate Concepts?):/i.test(snippet) ||
  isNormalizedGitStatusListing(snippet) ||
  /^User:\s*loops\/cron-self-heal\/runs\//i.test(snippet)
);
```

### Why require structured pairs

Do not replace the current matcher with something broad such as:

```ts
/\?\?\s+\S/
```

after normalization.

That could classify ordinary prose containing question marks as machine trace noise.

The token parser requires a known porcelain status token immediately followed by a path-like token
position. Two status/path pairs are enough for ordinary tracked listings; a single `??` or `!!`
pair is accepted because those markers are unambiguous.

If maintainers prefer a regex-only implementation, keep the same semantic constraint: match normalized
`status path status path` structure rather than newline boundaries.

## Regression must enter through real recall ingestion

Replace or supplement the raw-store fixture with a production-boundary case using
`recordShortTermRecalls()`.

The important input is a **real newline**, not a literal `\\n`:

```ts
const statusSnippet = " M tracked.md\n?? new.md";
const validSnippet = "User prefers short status updates.";

await recordShortTermRecalls({
  workspaceDir,
  query: "status and preference",
  nowMs,
  results: [
    {
      path: `memory/${DAY}-status.md`,
      startLine: 1,
      endLine: 2,
      score: 0.9,
      snippet: statusSnippet,
      source: "memory",
      provenance: {
        originClass: "owner",
        sessionKind: "interactive",
        observedAt: nowMs,
      },
    },
    {
      path: `memory/${DAY}-preference.md`,
      startLine: 1,
      endLine: 1,
      score: 0.9,
      snippet: validSnippet,
      source: "memory",
      provenance: {
        originClass: "owner",
        sessionKind: "interactive",
        observedAt: nowMs,
      },
    },
  ],
});
```

Create both source files first so any later live-source filtering remains representative.

Then read the actual recall store:

```ts
const recallStore = await shortTermTesting.readRecallStore(
  workspaceDir,
  new Date(nowMs).toISOString(),
);

const stored = Object.values(recallStore.entries);
expect(stored.map((entry) => entry.snippet)).toContain("M tracked.md ?? new.md");
```

That assertion proves the regression actually crossed the normalization boundary.

Now exercise the real candidate selectors:

```ts
const preview = previewRemDreaming({
  entries: stored,
  limit: 10,
  minPatternStrength: 0,
});

expect(preview.sourceEntryCount).toBe(1);
expect(preview.candidateKeys).toHaveLength(1);

const candidates = await rankShortTermPromotionCandidates({
  workspaceDir,
  nowMs,
  // use the same options/helper shape already used by the existing rankCandidates fixture
});

expect(candidates.map((entry) => entry.snippet)).toContain(validSnippet);
expect(candidates.map((entry) => entry.snippet)).not.toContain("M tracked.md ?? new.md");
```

Finally run REM dreaming and confirm only the legitimate preference produces reinforcement/promotion
signals.

## Negative controls

Add small direct predicate cases so the matcher does not become a prose filter:

```ts
expect(isDreamingTraceNoise({ path: "memory/x.md", snippet: "M tracked.md ?? new.md" })).toBe(true);
expect(isDreamingTraceNoise({ path: "memory/x.md", snippet: "?? new.md" })).toBe(true);

expect(
  isDreamingTraceNoise({
    path: "memory/x.md",
    snippet: "User prefers answers that explain why?? new ideas are welcome.",
  }),
).toBe(false);

expect(
  isDreamingTraceNoise({
    path: "memory/x.md",
    snippet: "M is the model size I usually choose.",
  }),
).toBe(false);
```

The exact negative-control prose can vary; the invariant is that punctuation and single-letter prose
must not be enough without porcelain structure.

## Real behavior proof

The reviewer separately asks for after-fix behavior through a real setup.

A compact proof can use a temporary workspace and the actual memory-core entrypoints:

```text
real memory search/recall result with multiline git-status snippet
→ recordShortTermRecalls
→ stored snippet observed as normalized single line
→ REM preview excludes it
→ deep promotion ranking excludes it

real preference recall in same batch
→ remains eligible
→ appears in REM candidate/reflection or promotion candidate
```

This does not require a live external model if the production dreaming selection/promotion entrypoint
used in the repository can be run deterministically; the important distinction from the current
fixture is that the trace enters through real recall recording rather than direct raw-store seeding.

A redacted terminal artifact should show:

```text
stored status: "M tracked.md ?? new.md"
REM status candidate: absent
deep status candidate: absent
stored preference: present
REM preference candidate: present
```

## Data-model note

This repair changes no persisted field, table, embedding metadata shape, config key, or migration.
It changes only filtering semantics over existing normalized snippet strings.

The automated data-model detector may flag dreaming/promotion files because they interact with
stored memory/vector data; that does not itself establish a stored-contract change.

## Acceptance

At minimum:

```bash
pnpm test extensions/memory-core/src/dreaming-phases.test.ts --maxWorkers=1
pnpm check:changed
git diff --check
```

Prefer also the memory-core database-worker shard used by the PR body.

Status: source-reviewed against `846873fde872558e8e8946a59977b5203ecd487c`; not executed by
`hippoley`.
