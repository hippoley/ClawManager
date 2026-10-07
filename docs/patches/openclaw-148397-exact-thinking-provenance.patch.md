# OpenClaw #148397 — exact thinking provenance patch skeleton

Target head: `7a88f9441f37d076b6b6c8dbf4529e812748aab7`

This version is grounded in the actual branch symbols and avoids widening the published
`OpenAICompletionsOptions` compatibility type.

## 1. Preserve reply-owner provenance

### `src/auto-reply/reply/reply-model-levels.ts`

```diff
 type ReplyModelLevels = {
   resolvedThinkLevel: ThinkLevel | undefined;
   resolvedReasoningLevel: ReasoningLevel;
+  thinkingExplicit: boolean;
 };
@@
-      return { resolvedThinkLevel, resolvedReasoningLevel };
+      return {
+        resolvedThinkLevel,
+        resolvedReasoningLevel,
+        thinkingExplicit: selection.thinkingExplicit,
+      };
```

The probe resolver already computes `hasExplicitThinkLevel` and passes it as
`selection.thinkingExplicit`, so no second provenance rule is needed.

### `src/auto-reply/reply/get-reply.ts`

```diff
-  const { resolvedThinkLevel, resolvedReasoningLevel } = await resolveRunModelLevels();
+  const {
+    resolvedThinkLevel,
+    resolvedReasoningLevel,
+    thinkingExplicit,
+  } = await resolveRunModelLevels();
@@
       directives,
       resolvedThinkLevel,
+      thinkingExplicit,
       resolvedReasoningLevel,
```

### `src/auto-reply/reply/get-reply-run.types.ts`

```diff
   resolvedThinkLevel: ThinkLevel | undefined;
+  /** True when the admitted turn/session/config selected thinking. */
+  thinkingExplicit?: boolean;
   resolvedFastMode?: FastMode;
```

Keep it optional on broad fixtures during rollout; the real reply resolver always supplies a boolean.

## 2. Put provenance beside the admitted think level

### `src/agents/command/shared-types.ts`

```diff
 export type AgentRunModelOptions = {
   ...
   thinkLevel?: ThinkLevel;
+  /** Selection provenance; false means known model-default inheritance. */
+  thinkingExplicit?: boolean;
   fastMode?: FastMode;
```

### `src/auto-reply/reply/get-reply-run-execute.ts`

The admitted followup run currently contains:

```ts
thinkingCatalog,
thinkLevel: resolvedThinkLevel,
thinkLevelOverride,
```

Change to:

```diff
       thinkingCatalog,
       thinkLevel: resolvedThinkLevel,
+      thinkingExplicit: params.thinkingExplicit,
       thinkLevelOverride,
```

This is the correct owner boundary because retries/fallbacks may remap `thinkLevel`, but they
must not reinterpret why the caller selected or did not select it.

## 3. Carry the bit into the session stream closure

### `src/agents/sessions/sdk.ts`

```diff
 export interface CreateAgentSessionOptions extends Omit<...> {
   ...
   thinkingLevel: ThinkingLevel;
+  /** Run-scoped selection provenance; not persisted in session state. */
+  thinkingExplicit?: boolean;
```

Avoid a new `SimpleStreamOptions` import entirely. Use the existing callback parameter type as
the local owner:

```diff
-      return modelRegistryRuntime.llmRuntime.streamSimple(modelResult, context, {
+      const providerOptions = {
         ...optionsLocal,
+        openclawThinkingExplicit: options.thinkingExplicit,
         apiKey: auth.apiKey,
         timeoutMs: optionsLocal?.timeoutMs ?? providerRetrySettings.timeoutMs,
         maxRetryDelayMs: optionsLocal?.maxRetryDelayMs ?? providerRetrySettings.maxRetryDelayMs,
         headers:
           attributionHeaders || auth.headers || optionsLocal?.headers
             ? { ...attributionHeaders, ...auth.headers, ...optionsLocal?.headers }
             : undefined,
-      });
+      } as NonNullable<typeof optionsLocal> & {
+        openclawThinkingExplicit?: boolean;
+      };
+      return modelRegistryRuntime.llmRuntime.streamSimple(modelResult, context, providerOptions);
```

That keeps this file's import surface unchanged and avoids touching `@openclaw/llm-core` public
types.

### `src/agents/embedded-agent-runner/run/attempt-session-prepare.ts`

```diff
     modelRegistry: attempt.modelRegistry,
     model: attempt.model,
     thinkingLevel: input.agentCoreThinkingLevel,
+    thinkingExplicit: attempt.thinkingExplicit,
     tools: sessionToolAllowlist,
```

No `AgentState`, transcript, session metadata, or persisted config change is needed.

## 4. Preserve the internal runtime field through the OpenAI simple wrapper

### `packages/ai/src/providers/openai-completions.ts`

Add a local narrow type:

```ts
type OpenAIThinkingProvenanceOptions = SimpleStreamOptions & {
  openclawThinkingExplicit?: boolean;
};
```

Then:

```diff
 export const streamSimpleOpenAICompletions: StreamFunction<
   "openai-completions",
   SimpleStreamOptions
 > = (model, context, options) => {
@@
   const toolChoice = (options as OpenAICompletionsOptions | undefined)?.toolChoice;
+  const openclawThinkingExplicit = (
+    options as OpenAIThinkingProvenanceOptions | undefined
+  )?.openclawThinkingExplicit;
 
-  return streamOpenAICompletions(model, context, {
+  const requestOptions = {
     ...base,
     reasoningEffort: clampedReasoning,
     toolChoice,
-  } satisfies OpenAICompletionsOptions);
+    openclawThinkingExplicit,
+  } as OpenAICompletionsOptions & {
+    openclawThinkingExplicit?: boolean;
+  };
+  return streamOpenAICompletions(model, context, requestOptions);
 };
```

Direct callers that already invoke `streamOpenAICompletions` can supply the same internal field
through a narrow intersection without changing the published compatibility type.

## 5. Gate the shared Qwen chat-template effort helper

### `packages/ai/src/transports/openai-transport-shared.ts`

The branch currently has:

```ts
export function resolveChatTemplateReasoningEffort(
  model: OpenAIModeModel,
  reasoning: { effort: string | undefined; thinkingEnabled: boolean | undefined },
): string | undefined {
  return reasoning.thinkingEnabled && readCompatReasoningEfforts(model.compat)?.length
    ? reasoning.effort
    : undefined;
}
```

Extend only its local reasoning input:

```diff
   reasoning: {
     effort: string | undefined;
     thinkingEnabled: boolean | undefined;
+    thinkingExplicit?: boolean;
   },
@@
-  return reasoning.thinkingEnabled && readCompatReasoningEfforts(model.compat)?.length
+  return (
+    reasoning.thinkingEnabled &&
+    reasoning.thinkingExplicit !== false &&
+    readCompatReasoningEfforts(model.compat)?.length
+  )
     ? reasoning.effort
     : undefined;
```

Three-state rollout semantics:

- `true` → selected by user/session/config; forward mapped effort;
- `false` → known inherited model default; omit nested effort;
- `undefined` → unmigrated caller; preserve current branch behavior.

## 6. Attach provenance to the reasoning object at the actual request owner

### `packages/ai/src/transports/openai-completions-params.ts`

The branch currently computes:

```ts
const reasoning = resolveOpenAIRequestReasoning(model, requestedEffort);
```

Replace with:

```ts
const thinkingExplicit = (
  options as
    | (OpenAICompletionsOptions & { openclawThinkingExplicit?: boolean })
    | undefined
)?.openclawThinkingExplicit;

const reasoning = {
  ...resolveOpenAIRequestReasoning(model, requestedEffort),
  thinkingExplicit,
};
```

Both existing calls now share the same provenance:

- managed Qwen chat-template branch calls
  `resolveChatTemplateReasoningEffort(model, reasoning)`;
- direct mode passes `reasoning` into
  `applyDirectCompletionsReasoningAndRouting(...)`, whose Qwen branch calls the same helper.

No duplicated provider-specific gate is required.

## 7. Regression matrix

Extend the existing OpenAI thinking contract coverage for **both managed and direct**:

1. known-unselected declared template:
   - `openclawThinkingExplicit: false`
   - `enable_thinking: true`
   - no nested `reasoning_effort`;

2. explicit selected high:
   - `openclawThinkingExplicit: true`
   - mapped nested effort remains present;

3. configured default:
   - owner stamps true;
   - mapped nested effort remains present;

4. off:
   - `enable_thinking: false`
   - nested effort absent;

5. unknown/unmigrated caller:
   - provenance undefined;
   - current candidate behavior preserved during rollout.

## 8. Non-reply owners — exact stamp points

Do not guess from the resolved effort value.

### CLI / agent exec

`src/agents/command/model-selection.ts` already computes:

```ts
const immutableThinkLevel = params.requestedThinkLevel ?? configuredThinkLevel;
const primaryConfiguredThinkLevel =
  immutableThinkLevel ??
  resolveConfiguredThinkingDefault({
    cfg: params.cfg,
    agentId: params.sessionAgentId,
    provider,
    model,
  });
```

This is the exact provenance boundary. `resolveConfiguredThinkingDefault()` reads real caller/
configuration sources, while `resolveThinkingSelection(...)` is the later model-capability/default
resolver.

Return one additional fact beside `effectiveTurnThinkLevel`:

```diff
     immutableThinkLevel,
     effectiveTurnThinkLevel: primaryThinking.requestedLevel,
+    thinkingExplicit: primaryConfiguredThinkLevel !== undefined,
     sessionFile,
```

Then carry that fact into the `AgentRunModelOptions` object for the admitted CLI/agent-exec run.

### Subagent spawn

Subagent ownership is more precise than simple parent inheritance.

`src/agents/subagents/spawn/subagent-spawn-thinking.ts` already distinguishes:

1. explicit `params.thinkingOverrideRaw`;
2. requester-agent `subagents.thinking`;
3. target-agent `subagents.thinking`;
4. global `agents.defaults.subagents.thinking`;
5. caller/requester persisted/active `callerThinkingRaw`.

The current owner:

```ts
const resolvedThinkingDefaultRaw =
  requesterAgentConfig?.subagents?.thinking ??
  targetAgentConfig?.subagents?.thinking ??
  cfg.agents?.defaults?.subagents?.thinking;

const overrideCandidateRaw = thinkingOverrideRaw || resolvedThinkingDefaultRaw;
```

can return provenance directly:

```diff
     return {
       status: "ok" as const,
       thinkingOverride: normalizedThinking,
+      thinkingExplicit: true,
       initialSessionPatch: {
         thinkingLevel: normalizedThinking,
       },
     };
@@
   return {
     status: "ok" as const,
     thinkingOverride: undefined,
+    thinkingExplicit: normalizedThinking !== undefined,
     initialSessionPatch: normalizedThinking ? { thinkingLevel: normalizedThinking } : {},
   };
```

Why `normalizedThinking !== undefined` is correct in the inherited branch:

- `callerThinkingRaw` comes from the active requester's thinking level or persisted requester
  preference;
- it is therefore an actual selected/session preference, not a child-model capability fallback.

`src/agents/subagents/spawn/subagent-spawn-child-plan.ts` already passes
`thinkingOverrideRaw` and `callerThinkingRaw` into this owner. Carry the returned
`thinkingExplicit` beside the child run's selected thinking level when constructing the child
session/run.

This avoids blindly inheriting the parent's provenance when the child has its own
`subagents.thinking` override.

### Cron

Keep the same rule already identified for cron: stamp true when its pre-resolution
`requestedThinkLevel` exists; false only when the final level is supplied solely by model
capability/default resolution.

### Unknown internal callers

Leave `thinkingExplicit` undefined until their owner is audited. The request helper deliberately
treats undefined as current candidate behavior, so migrating Gateway/CLI/subagent paths does not
silently suppress effort elsewhere.

## Focused validation

```bash
node scripts/run-vitest.mjs src/auto-reply/reply/reply-model-levels.test.ts --run
node scripts/run-vitest.mjs src/agents/openai-thinking-contract.test.ts --run
node scripts/run-vitest.mjs packages/ai/src/transports/openai-completions-params.test.ts --run
pnpm tsgo:core
git diff --check
```

Status: source-reviewed against `7a88f9441f37d076b6b6c8dbf4529e812748aab7`; not executed.


## Self-audit refresh (2026-10-07)

- Current `sdk.ts` does not import `SimpleStreamOptions`; the revised closure uses
  `NonNullable<typeof optionsLocal>` and requires no new public/runtime type import.
- Current `resolveEmbeddedModelSelection()` exposes the exact CLI provenance boundary:
  `primaryConfiguredThinkLevel !== undefined` before model fallback selection.
- Current subagent thinking owner has an explicit precedence chain and can stamp provenance without
  inspecting the final resolved effort.
- The shared `resolveChatTemplateReasoningEffort()` helper exists on this PR head and is called by
  both managed and direct Chat Completions paths, so one provenance gate remains the correct
  serialization owner.
