# OpenClaw #154702 — exact current-main patch

This patch is adapted to current main's split Google imports (`./http-request.js` + `./model-id.js`) while preserving the reviewed timeout semantics.

```diff
diff --git a/extensions/google/image-generation-provider.ts b/extensions/google/image-generation-provider.ts
--- a/extensions/google/image-generation-provider.ts
+++ b/extensions/google/image-generation-provider.ts
@@
+import { buildTimeoutAbortSignal } from "openclaw/plugin-sdk/extension-shared";
 import {
   generatedImageAssetFromBase64,
   resolveInlineImageJsonResponseMaxBytes,
@@
 export function buildGoogleImageGenerationProvider(): ImageGenerationProvider {
   return {
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
-
-      const model = normalizeGoogleModelId(req.model?.trim() || DEFAULT_GOOGLE_IMAGE_MODEL);
-      const { baseUrl, allowPrivateNetwork, headers, dispatcherPolicy } =
-        resolveGoogleGenerativeAiHttpRequestConfig({
-          apiKey: auth.apiKey,
-          baseUrl: req.cfg?.models?.providers?.google?.baseUrl,
-          request: sanitizeConfiguredModelProviderRequest(
-            req.cfg?.models?.providers?.google?.request,
-          ),
-          capability: "image",
-          transport: "http",
-        });
-      const imageConfig = mapSizeToImageConfig(req.size);
-      const inputParts = (req.inputImages ?? []).map((image) => ({
-        inlineData: {
-          mimeType: image.mimeType,
-          data: image.buffer.toString("base64"),
-        },
-      }));
-      const resolvedImageConfig = {
-        ...imageConfig,
-        ...(req.aspectRatio?.trim() ? { aspectRatio: req.aspectRatio.trim() } : {}),
-        ...(req.resolution ? { imageSize: req.resolution } : {}),
-      };
-
-      const { response: res, release } = await postJsonRequest({
-        url: `${baseUrl}/models/${model}:generateContent`,
-        headers,
-        body: {
-          contents: [
-            {
-              role: "user",
-              parts: [...inputParts, { text: req.prompt }],
-            },
-          ],
-          generationConfig: {
-            responseModalities: ["TEXT", "IMAGE"],
-            ...(Object.keys(resolvedImageConfig).length > 0
-              ? { imageConfig: resolvedImageConfig }
-              : {}),
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
+
+        const model = normalizeGoogleModelId(req.model?.trim() || DEFAULT_GOOGLE_IMAGE_MODEL);
+        const { baseUrl, allowPrivateNetwork, headers, dispatcherPolicy } =
+          resolveGoogleGenerativeAiHttpRequestConfig({
+            apiKey: auth.apiKey,
+            baseUrl: req.cfg?.models?.providers?.google?.baseUrl,
+            request: sanitizeConfiguredModelProviderRequest(
+              req.cfg?.models?.providers?.google?.request,
+            ),
+            capability: "image",
+            transport: "http",
+          });
+        const imageConfig = mapSizeToImageConfig(req.size);
+        const inputParts = (req.inputImages ?? []).map((image) => ({
+          inlineData: {
+            mimeType: image.mimeType,
+            data: image.buffer.toString("base64"),
+          },
+        }));
+        const resolvedImageConfig = {
+          ...imageConfig,
+          ...(req.aspectRatio?.trim() ? { aspectRatio: req.aspectRatio.trim() } : {}),
+          ...(req.resolution ? { imageSize: req.resolution } : {}),
+        };
+
+        const { response: res, release } = await postJsonRequest({
+          url: `${baseUrl}/models/${model}:generateContent`,
+          headers,
+          body: {
+            contents: [
+              {
+                role: "user",
+                parts: [...inputParts, { text: req.prompt }],
+              },
+            ],
+            generationConfig: {
+              responseModalities: ["TEXT", "IMAGE"],
+              ...(Object.keys(resolvedImageConfig).length > 0
+                ? { imageConfig: resolvedImageConfig }
+                : {}),
+            },
           },
-        },
-        timeoutMs: req.timeoutMs ?? DEFAULT_IMAGE_TIMEOUT_MS,
-        fetchFn: fetch,
-        pinDns: false,
-        allowPrivateNetwork,
-        ssrfPolicy: req.ssrfPolicy,
-        dispatcherPolicy,
-      });
-
-      try {
-        await assertOkOrThrowHttpError(res, "Google image generation failed");
+          timeoutMs: req.timeoutMs ?? DEFAULT_IMAGE_TIMEOUT_MS,
+          ...(signal ? { signal } : {}),
+          fetchFn: fetch,
+          pinDns: false,
+          allowPrivateNetwork,
+          ssrfPolicy: req.ssrfPolicy,
+          dispatcherPolicy,
+        });
 
-        const payload = await readProviderJsonResponse(res, "google.image-generation", {
-          maxBytes: resolveInlineImageJsonResponseMaxBytes(
-            GOOGLE_MAX_IMAGE_RESULTS,
-            resolveGeneratedMediaMaxBytes(req.cfg, "image"),
-          ),
-        });
-        const images: GeneratedImageAsset[] = [];
-        for (const part of googleResponseParts(payload)) {
-          const inline = googleInlineDataFromPart(part);
-          if (!inline) {
-            continue;
-          }
-          const data = normalizeOptionalString(inline.data);
-          if (!data) {
-            throw new Error(GOOGLE_IMAGE_MALFORMED_RESPONSE);
-          }
-          const standardData = toStandardGoogleProviderBase64(data);
-          if (!standardData) {
-            throw new Error(GOOGLE_IMAGE_MALFORMED_RESPONSE);
-          }
-          const image = generatedImageAssetFromBase64({
-            base64: standardData,
-            index: images.length,
-            mimeType:
-              normalizeOptionalString(inline.mimeType) ??
-              normalizeOptionalString(inline.mime_type) ??
-              DEFAULT_OUTPUT_MIME,
+        try {
+          await assertOkOrThrowHttpError(res, "Google image generation failed");
+
+          const payload = await readProviderJsonResponse(res, "google.image-generation", {
+            maxBytes: resolveInlineImageJsonResponseMaxBytes(
+              GOOGLE_MAX_IMAGE_RESULTS,
+              resolveGeneratedMediaMaxBytes(req.cfg, "image"),
+            ),
           });
-          if (!image) {
-            throw new Error(GOOGLE_IMAGE_MALFORMED_RESPONSE);
+          const images: GeneratedImageAsset[] = [];
+          for (const part of googleResponseParts(payload)) {
+            const inline = googleInlineDataFromPart(part);
+            if (!inline) {
+              continue;
+            }
+            const data = normalizeOptionalString(inline.data);
+            if (!data) {
+              throw new Error(GOOGLE_IMAGE_MALFORMED_RESPONSE);
+            }
+            const standardData = toStandardGoogleProviderBase64(data);
+            if (!standardData) {
+              throw new Error(GOOGLE_IMAGE_MALFORMED_RESPONSE);
+            }
+            const image = generatedImageAssetFromBase64({
+              base64: standardData,
+              index: images.length,
+              mimeType:
+                normalizeOptionalString(inline.mimeType) ??
+                normalizeOptionalString(inline.mime_type) ??
+                DEFAULT_OUTPUT_MIME,
+            });
+            if (!image) {
+              throw new Error(GOOGLE_IMAGE_MALFORMED_RESPONSE);
+            }
+            images.push(image);
           }
-          images.push(image);
-        }
-
-        if (images.length === 0) {
-          throw new Error("Google image generation response missing image data");
-        }
 
-        return {
-          images,
-          model,
-        };
+          if (images.length === 0) {
+            throw new Error("Google image generation response missing image data");
+          }
+
+          return { images, model };
+        } finally {
+          await release();
+        }
       } finally {
-        await release();
+        cleanup();
       }
     },
   };
 }
```

## Exact regression additions

Append the three already-reviewed cases from the candidate branch to
`extensions/google/image-generation-provider.test.ts`:

- `aborts credential preparation when it exceeds the request timeout`
- `forwards the operation signal into the image HTTP request`
- `does not apply an abort signal to credential preparation when timeout is omitted`

Those cases already match the current test seam (`providerAuthRuntime` and `providerHttp` spies), so no current-main production seam is needed.

## Focused validation

```bash
node scripts/run-vitest.mjs extensions/google/image-generation-provider.test.ts --run
pnpm exec oxlint extensions/google/image-generation-provider.ts extensions/google/image-generation-provider.test.ts
pnpm exec oxfmt --check extensions/google/image-generation-provider.ts extensions/google/image-generation-provider.test.ts
```

Compatibility invariant:

```text
req.timeoutMs supplied
  => one signal spans credential prep + HTTP

req.timeoutMs omitted
  => credential prep keeps previous unbounded behavior
  => HTTP keeps DEFAULT_IMAGE_TIMEOUT_MS
```

Status: exact source patch adapted to current main; not executed in this environment.


## Self-audit / portability check (2026-10-07)

Verified against current main and the candidate branch:

- `buildTimeoutAbortSignal` accepts `timeoutMs?: number`; passing `req.timeoutMs` directly is type-correct and intentionally yields no operation signal when omitted.
- `openclaw/plugin-sdk/extension-shared` is the established extension import path for `buildTimeoutAbortSignal`.
- `resolveApiKeyForProvider` accepts `signal`.
- `postJsonRequest` accepts `signal`.
- Current-main `extensions/google/image-generation-provider.test.ts` still imports both `providerAuthRuntime` and `providerHttp`, so the candidate's three fake-timer regressions can be appended without introducing a new test seam.
- Candidate branch `fix/google-generation-credential-abort-signal-080` uses the same three regression names documented above.

This confirms the patch is portable at the current API/test-seam level. It remains source-reviewed, not locally executed.
