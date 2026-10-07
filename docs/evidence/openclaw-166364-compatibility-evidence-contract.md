# OpenClaw #166364 — compatibility evidence contract

Purpose: give maintainers a concrete way to dispose of the stale stored-data / migration blocker without inventing a migration story for a behavior-preserving refactor.

## Source finding

The latest ClawSweeper review already says:

- previous migration-proof blocker is unsupported by the diff;
- repair writes, runtime-pin precedence, shielding, and serialized fields remain unchanged;
- no material authority change or stored-data contract change was identified.

The remaining checklist item is therefore best handled as an equivalence proof.

## Flagged Doctor/runtime-policy files

The detector flags:

- `src/commands/doctor/shared/codex-route-compaction-repair.ts`
- `src/commands/doctor/shared/codex-route-config-repair.ts`
- `src/commands/doctor/shared/codex-route-runtime-policy.ts`
- `src/flows/doctor-tool-schema-projection.ts`
- `src/flows/doctor-tool-schema-runtime.ts`

Current source comparison shows the candidate refactor does not add a persisted field owner.

Examples:

### Runtime policy

Base passes `isDefaults?: boolean` into `ensureCodexRuntimePolicy`.

Candidate removes that redundant input and derives the same condition from the existing owner:

```ts
const isDefaults = params.agentPath === "agents.defaults";
```

The mutation owner remains `setModelRuntimePolicy(...)`, and the same `agent.models[modelRef].agentRuntime.id` field is written.

### Tool-schema runtime/projection

Candidate changes repeated explicit forwarding such as:

```ts
return collectNormalizedToolSchemaFindings({
  agentId: params.agentId,
  tools: activeBundleTools,
  cfg: params.cfg,
  workspaceDir: params.workspaceDir,
  modelRef: params.modelRef,
  model: params.model,
  normalizationFailureFinding,
});
```

to equivalent forwarding:

```ts
return collectNormalizedToolSchemaFindings({
  ...params,
  tools: activeBundleTools,
  normalizationFailureFinding,
});
```

This changes call-site projection style, not stored output shape.

## Recommended compatibility proof

Use existing Codex-route repair fixtures and compare base vs candidate outputs for the same inputs.

For each fixture:

1. clone the input config twice;
2. run the existing repair owner on base;
3. run the same repair owner on candidate;
4. compare normalized JSON output;
5. compare `changes[]` entries where wording is contractually relevant;
6. assert no extra keys appear under:
   - `agents.defaults.models[*]`
   - `agents.entries[*].models[*]`
   - `agents.list[*].models[*]`
   - provider/model runtime policy fields.

Representative cases:

- default-agent Codex route with no explicit runtime pin;
- explicit listed agent referencing the same canonical model;
- provider-level non-default runtime pin;
- provider-model runtime pin;
- existing explicit `agentRuntime.id`;
- legacy Codex ref requiring canonical repair;
- compaction / memory-flush model references.

## Exact assertion shape

```ts
expect(normalizeConfig(candidateResult.cfg)).toEqual(normalizeConfig(baseResult.cfg));
expect(listPersistedKeys(candidateResult.cfg)).toEqual(listPersistedKeys(baseResult.cfg));
```

For the specific `isDefaults` refactor, also assert:

```ts
expect(candidateDefaultRepair).toEqual(baseDefaultRepair);
expect(candidateExplicitAgentRepair).toEqual(baseExplicitAgentRepair);
```

## Tool-schema compatibility

For the two tool-schema files, compare emitted `HealthFinding[]` for existing fixtures:

```ts
expect(candidateFindings).toEqual(baseFindings);
```

There is no config write in this path, so a migration test is not the appropriate contract.

## Maintainer conclusion if equivalence passes

A sufficient disposition would be:

> Compatibility checked by behavior-equivalence replay across the existing repair/projection owners. Candidate and base emit the same persisted config shape and health findings for representative default-agent, explicit-agent, runtime-pin, and legacy-route fixtures. No new persisted keys, removed keys, migration, or backfill path was introduced.

Status: source-reviewed against reviewed head `4af8770dfe545f941d73934a87bd08ea2c175913`; not executed in this environment.
