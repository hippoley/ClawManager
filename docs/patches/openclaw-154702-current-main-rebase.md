# OpenClaw #154702 — current-main rebase patch

The PR's behavior is still missing on current `main`: Google image generation resolves credentials before any operation-level abort signal exists, so an explicit `timeoutMs` does not bound credential preparation.

The merge conflict is mostly code drift:
- current main imports `resolveGoogleGenerativeAiHttpRequestConfig` from `./http-request.js`;
- current main imports `normalizeGoogleModelId` from `./model-id.js`;
- the provider body otherwise still has the same pre-fix ownership boundary.

## Minimal current-main integration

```diff
diff --git a/extensions/google/image-generation-provider.ts b/extensions/google/image-generation-provider.ts
--- a/extensions/google/image-generation-provider.ts
+++ b/extensions/google/image-generation-provider.ts
@@
+import { buildTimeoutAbortSignal } from "openclaw/plugin-sdk/extension-shared";
 import {
   generatedImageAssetFromBase64,
@@
     ...createGoogleImageGenerationProviderMetadata(),
     isConfigured: (ctx) => isProviderApiKeyConfigured({ provider: "google", ...ctx }),
     async generateImage(req) {
-      const auth = await resolveApiKeyForProvider({
-        provider: "google",
-        cfg: req.cfg,
-        agentDir: req.agentDir,
-        store: req.authStore,
+      const { signal, cleanup } = buildTimeoutAbortSignal({
+        timeoutMs: req.timeoutMs,
+        operation: "Google image generation",
       });
-      if (!auth.apiKey) {
-        throw new Error("Google API key missing");
-      }
+      try {
+        signal?.throwIfAborted();
+        const auth = await resolveApiKeyForProvider({
+          provider: "google",
+          cfg: req.cfg,
+          agentDir: req.agentDir,
+          store: req.authStore,
+          ...(signal ? { signal } : {}),
+        });
+        signal?.throwIfAborted();
+        if (!auth.apiKey) {
+          throw new Error("Google API key missing");
+        }

-      const model = normalizeGoogleModelId(req.model?.trim() || DEFAULT_GOOGLE_IMAGE_MODEL);
+        const model = normalizeGoogleModelId(req.model?.trim() || DEFAULT_GOOGLE_IMAGE_MODEL);
         // keep the current-main http-request/model-id imports unchanged
         // keep current request construction unchanged

-      const { response: res, release } = await postJsonRequest({
+        const { response: res, release } = await postJsonRequest({
           // existing current-main request fields
           timeoutMs: req.timeoutMs ?? DEFAULT_IMAGE_TIMEOUT_MS,
+          ...(signal ? { signal } : {}),
           fetchFn: fetch,
           pinDns: false,
           allowPrivateNetwork,
           ssrfPolicy: req.ssrfPolicy,
           dispatcherPolicy,
-      });
+        });

-      try {
-        // existing response parsing
-      } finally {
-        await release();
+        try {
+          // existing response parsing unchanged
+        } finally {
+          await release();
+        }
+      } finally {
+        cleanup();
       }
     },
   };
 }
```

## Important compatibility point

Do **not** replace `req.timeoutMs` with `req.timeoutMs ?? DEFAULT_IMAGE_TIMEOUT_MS` when creating the operation signal.

That preserves current behavior:
- explicit timeout -> credential preparation + HTTP share one absolute budget;
- omitted timeout -> credential preparation remains unbounded as before;
- HTTP still keeps its existing `DEFAULT_IMAGE_TIMEOUT_MS` fallback.

## Regression set

Reapply the three existing candidate regressions onto the current-main test file:

1. explicit timeout aborts stalled credential preparation;
2. the same operation signal is forwarded into `postJsonRequest`;
3. omitted timeout does not attach an abort signal to credential preparation.

The current-main test file already mocks `providerAuthRuntime` and `providerHttp`, so no new test seam is needed.

## Suggested focused validation

```bash
pnpm test extensions/google/image-generation-provider.test.ts
pnpm exec oxlint extensions/google/image-generation-provider.ts extensions/google/image-generation-provider.test.ts
pnpm exec oxfmt --check extensions/google/image-generation-provider.ts extensions/google/image-generation-provider.test.ts
```

## Why this is safe to rebase narrowly

The PR already has strong real-behavior evidence:
- real credential owner;
- real undici transport;
- local HTTPS stall;
- bounded credential cancellation;
- bounded HTTP cancellation;
- successful recovery;
- omitted-timeout compatibility.

Current-main inspection shows no new owner has superseded this fix. The conflict is source-layout drift, not a semantic replacement.

Status: source-reviewed against current main on 2026-10-07; not executed in this environment.
