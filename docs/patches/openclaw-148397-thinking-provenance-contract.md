# OpenClaw #148397 — preserve unselected Qwen template defaults via thinking provenance

Target head reviewed: `7a88f9441f37d076b6b6c8dbf4529e812748aab7`

## Root cause on the current head

The current branch correctly forwards declared Qwen chat-template efforts, but it still loses the distinction between:

```text
explicit/configured high
vs
high chosen only because the model profile defaulted there
```

That distinction already exists in the reply owner:

```ts
const thinkingExplicitlySet =
  thinkingLevelOverride !== undefined ||
  directives.thinkLevel !== undefined ||
  sessionThinkLevel !== undefined ||
  configuredThinkingDefault !== undefined ||
  modelState.hasConfiguredThinkingDefault === true;
```

But `ReplyModelLevelResolver` currently returns only:

```ts
{
  resolvedThinkLevel,
  resolvedReasoningLevel,
}
```

so the provenance bit is dropped before the run/provider boundary.

The transport then receives only a resolved effort:

```ts
resolveChatTemplateReasoningEffort(model, reasoning)
```

where `reasoning` contains only `effort` and `thinkingEnabled`. At that point `"high"` cannot safely be interpreted as explicit or inherited.

## Compatibility rule

For a declared `qwen-chat-template` model:

```text
no user/session/configured thinking choice
  => keep enable_thinking behavior
  => OMIT nested reasoning_effort
  => backend chat template keeps its own default

explicit or configured thinking choice
  => map through supportedReasoningEfforts / reasoningEffortMap
  => send nested reasoning_effort

off
  => enable_thinking false
  => no nested reasoning_effort
```

Do not infer provenance from the resolved effort string.

## Minimal implementation shape

### 1. Preserve provenance in the existing resolver

`src/auto-reply/reply/reply-model-levels.ts`

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

For probe recomputation, keep using the already-computed `hasExplicitThinkLevel`; the returned bit naturally follows the replacement selection.

### 2. Carry it through the prepared reply

`src/auto-reply/reply/get-reply.ts`

```diff
-  const { resolvedThinkLevel, resolvedReasoningLevel } = await resolveRunModelLevels();
+  const {
+    resolvedThinkLevel,
+    resolvedReasoningLevel,
+    thinkingExplicit,
+  } = await resolveRunModelLevels();
@@
     runPreparedReply({
       ...
       resolvedThinkLevel,
       resolvedReasoningLevel,
+      thinkingExplicit,
       ...
     })
```

`src/auto-reply/reply/get-reply-run.types.ts`

```diff
   resolvedThinkLevel: ThinkLevel | undefined;
+  /** True when the turn/session/config selected thinking rather than inheriting a model default. */
+  thinkingExplicit: boolean;
   resolvedReasoningLevel: ReasoningLevel;
```

### 3. Put the bit beside `thinkLevel` at the run owner

`src/agents/command/shared-types.ts`

```diff
 export type AgentRunModelOptions = {
   ...
   thinkLevel?: ThinkLevel;
+  /** Provenance for thinkLevel; never infer this from the level value. */
+  thinkingExplicit?: boolean;
   fastMode?: FastMode;
```

When `executePreparedReplyRun` / its run-admission builder constructs the run object, carry:

```ts
thinkingExplicit: params.thinkingExplicit,
```

Fallback candidates may recompute `candidateThinkLevel`, but they should retain the admitted turn's `thinkingExplicit` bit. A fallback model change does not retroactively turn an inherited default into a user/config choice.

### 4. Carry it into the session through the existing stream closure

The smallest implementation does **not** need to modify `AgentState`, `AgentOptions`,
`AgentLoopConfig`, or `llm-core/SimpleStreamOptions`.

`createAgentSession()` already supplies a custom `streamFn` closure to `Agent`:

```ts
streamFn: async (modelResult, context, optionsLocal) => {
  ...
  return modelRegistryRuntime.llmRuntime.streamSimple(modelResult, context, {
    ...optionsLocal,
    ...
  });
}
```

That closure can capture one new run/session option directly.

In `src/agents/sessions/sdk.ts`, extend only `CreateAgentSessionOptions`:

```diff
 export interface CreateAgentSessionOptions extends Omit<
   AgentSessionConfig,
   "agent" | "cwd" | "extensionRunnerRef" | "allowedToolNames"
 > {
+  /** Whether the admitted turn/session/config explicitly selected thinking. */
+  thinkingExplicit?: boolean;
   ...
 }
```

In `prepareEmbeddedAttemptAgentSession`:

```diff
 const sessionOptions: CreateAgentSessionOptions = {
   ...
   thinkingLevel: input.agentCoreThinkingLevel,
+  thinkingExplicit: attempt.thinkingExplicit,
   ...
 };
```

Then use a narrow local provider-options extension in the `createAgentSession()` stream closure:

```ts
type ThinkingProvenanceStreamOptions = SimpleStreamOptions & {
  thinkingExplicit?: boolean;
};
```

and:

```diff
- return modelRegistryRuntime.llmRuntime.streamSimple(modelResult, context, {
+ const providerOptions: ThinkingProvenanceStreamOptions = {
    ...optionsLocal,
+   thinkingExplicit: options.thinkingExplicit,
    apiKey: auth.apiKey,
    timeoutMs: optionsLocal?.timeoutMs ?? providerRetrySettings.timeoutMs,
    maxRetryDelayMs: optionsLocal?.maxRetryDelayMs ?? providerRetrySettings.maxRetryDelayMs,
    headers:
      attributionHeaders || auth.headers || optionsLocal?.headers
        ? { ...attributionHeaders, ...auth.headers, ...optionsLocal?.headers }
        : undefined,
- });
+ };
+ return modelRegistryRuntime.llmRuntime.streamSimple(
+   modelResult,
+   context,
+   providerOptions,
+ );
```

Because TypeScript is structurally typed, the extended object remains assignable to
`SimpleStreamOptions`; the extra property is retained at runtime for the OpenAI wrapper.

This is strictly run-scoped request provenance. It never enters:
- `AgentState`;
- the transcript;
- session metadata;
- model catalog state;
- persisted config.

### 5. Read the provenance only in the OpenAI/Qwen wrapper

Do **not** add `thinkingExplicit` to `packages/llm-core/src/types.ts`.

In `packages/ai/src/providers/openai-completions.ts`, read the runtime extra field through the
same narrow intersection:

```ts
type OpenAIThinkingProvenanceOptions = SimpleStreamOptions & {
  thinkingExplicit?: boolean;
};

const thinkingExplicit =
  (options as OpenAIThinkingProvenanceOptions | undefined)?.thinkingExplicit;
```

Pass it forward as an internal extra property when invoking the concrete
`OpenAICompletionsOptions` path:

```ts
const requestOptions = {
  ...base,
  reasoningEffort: clampedReasoning,
  toolChoice,
  thinkingExplicit,
} as OpenAICompletionsOptions & { thinkingExplicit?: boolean };

return streamOpenAICompletions(model, context, requestOptions);
```

At the transport owner, read that same narrow extension and combine it with the existing
`resolveOpenAIRequestReasoning(...)` result:

```ts
const reasoning = {
  ...resolveOpenAIRequestReasoning(...),
  thinkingExplicit:
    (options as (OpenAICompletionsOptions & { thinkingExplicit?: boolean }) | undefined)
      ?.thinkingExplicit,
};
```

Then tighten only the Qwen chat-template helper:

```ts
export function resolveChatTemplateReasoningEffort(
  model: OpenAIModeModel,
  reasoning: {
    effort: string | undefined;
    thinkingEnabled: boolean | undefined;
    thinkingExplicit?: boolean;
  },
): string | undefined {
  return reasoning.thinkingEnabled &&
    reasoning.thinkingExplicit === true &&
    readCompatReasoningEfforts(model.compat)?.length
    ? reasoning.effort
    : undefined;
}
```

Both managed and direct Chat Completions paths already consume this shared helper, so one
provenance-aware gate covers both transports.

### 6. Exact run-object hop

The admitted reply run is assembled in
`src/auto-reply/reply/get-reply-run-execute.ts`, where `thinkLevel` already enters
`followupRun.run`:

```ts
run: {
  ...
  thinkingCatalog,
  thinkLevel: resolvedThinkLevel,
  ...
}
```

Add the provenance beside the value:

```diff
   thinkingCatalog,
   thinkLevel: resolvedThinkLevel,
+  thinkingExplicit: params.thinkingExplicit,
   thinkLevelOverride,
```

and add `thinkingExplicit?: boolean` beside `thinkLevel?: ThinkLevel` in
`AgentRunModelOptions`.

This makes `attempt.thinkingExplicit` available automatically because
`RunEmbeddedAgentParams` already intersects `AgentRunModelOptions`.

Fallback candidates may recompute `candidateThinkLevel`, but the admitted run's
`thinkingExplicit` bit should remain unchanged: changing model candidates does not turn an
inherited default into an operator/session/config choice.

## Important non-goals

Do **not**:

- remove declared Qwen effort forwarding;
- infer explicitness from `effort === "high"`;
- change ordinary binary `qwen` behavior;
- change undeclared `qwen-chat-template` behavior;
- persist the provenance bit;
- add a migration or config key;
- change vLLM's binary enable/disable control.

## Regression matrix

Extend `src/agents/openai-thinking-contract.test.ts` over both `managed` and `direct`:

### A. Unselected declared template

A declared model with no per-turn/session/configured thinking choice:

```ts
expect(payload.chat_template_kwargs).toMatchObject({
  enable_thinking: true,
});
expect(payload).not.toHaveProperty("chat_template_kwargs.reasoning_effort");
```

This is the upgrade-compatibility control and should use a conceptual template whose server default differs from mapped high.

### B. Explicit selected level

Existing selected-level cases remain green:

```text
minimal -> low
low     -> low
medium  -> medium
high    -> xhigh
xhigh   -> xhigh
max     -> xhigh
```

### C. Configured default

A configured agent/model thinking default counts as selected policy and still sends the mapped nested effort.

The existing `thinkingExplicitlySet` owner already defines this semantic; do not invent a second rule in the transport.

### D. Off

Existing off case remains:

```ts
enable_thinking: false
no reasoning_effort
```

### E. Undeclared template

Existing binary-only test remains unchanged.

## Why this is smaller than changing thinking defaults

Changing `buildBaseThinkingProfile(... "high")` alone cannot solve the transport compatibility issue safely:

- a model may still resolve to `high` from another default owner;
- the transport still cannot distinguish inherited vs selected;
- explicit `high` and inherited `high` serialize identically before provenance is added.

One boolean provenance bit preserves the existing thinking resolution while giving the serialization boundary the information it actually needs.

## Suggested focused validation

```bash
node scripts/run-vitest.mjs src/auto-reply/reply/reply-model-levels.test.ts --run
node scripts/run-vitest.mjs src/agents/openai-thinking-contract.test.ts --run
node scripts/run-vitest.mjs src/agents/sessions/sdk.test.ts --run
```

Then rerun the existing real Qwen/vLLM proof for selected effort forwarding. For upgrade proof, use a declared template whose server default differs from mapped high and show that an unselected turn omits `reasoning_effort`.

Status: source-reviewed against head `7a88f9441f37d076b6b6c8dbf4529e812748aab7`; not executed.
