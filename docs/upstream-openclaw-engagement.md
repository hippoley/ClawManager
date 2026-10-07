# Upstream OpenClaw engagement

This page records upstream OpenClaw work that is being tracked, reviewed, reproduced, or contributed from the `hippoley` account and related local work. It is deliberately provenance-first: an upstream PR is only called **authored** when the public GitHub author is `hippoley`. Work authored by other contributors is listed as upstream context, not claimed as my PR.

Snapshot: 2026-10-07.

## Attribution rules

- **Authored** — PR / issue / review / comment is publicly attributable to `hippoley`.
- **Contributed evidence** — reproduction, benchmark, runtime proof, compatibility proof, review, or other material evidence is publicly attributable to `hippoley`.
- **Upstream incorporated** — related work was superseded or reimplemented upstream; this is not presented as a merged `hippoley` PR unless GitHub attribution proves it.
- **Tracked upstream** — useful engineering context only. No authorship claim.

## Merged upstream context

| Upstream item | Result | Attribution note |
| --- | --- | --- |
| [#166362](https://github.com/openclaw/openclaw/pull/166362) | Merged | Authored by `shakkernerd`; tracked upstream, not claimed as a `hippoley` PR. |
| [#165818](https://github.com/openclaw/openclaw/pull/165818) | Merged; closes [#139151](https://github.com/openclaw/openclaw/issues/139151) and supersedes [#139818](https://github.com/openclaw/openclaw/pull/139818) | Authored by `steipete`; #139818 was authored by `vortexopenclaw`. This chain is kept as upstream context until a public `hippoley` artifact proves direct contribution. |

## Active blocker-specific work

| Upstream item | Current state | Next useful action |
| --- | --- | --- |
| [#166365](https://github.com/openclaw/openclaw/pull/166365) | Open | Maintainer decision / review. Avoid duplicate implementation unless a concrete finding appears. |
| [#166364](https://github.com/openclaw/openclaw/pull/166364) | Open — **contributed evidence** | [hippoley compatibility-proof note](https://github.com/openclaw/openclaw/pull/166364#issuecomment-6029001497): narrowed the blocker to repair-boundary behavior equivalence and proposed base-vs-candidate persisted-output replay instead of an invented migration narrative. |
| [#166361](https://github.com/openclaw/openclaw/pull/166361) | ✅ **Merged — blocker resolution contribution** | [hippoley CI-cost note](https://github.com/openclaw/openclaw/pull/166361#issuecomment-6029161635) surfaced the missing CI-cost record. The author triggered a fresh review immediately afterward; Revision 6 cleared the blocker, and the PR was subsequently **merged on 2026-10-07**. |
| [#154735](https://github.com/openclaw/openclaw/pull/154735) | Open — **contributed current-main patch** | [hippoley current-main integration patch](https://github.com/hippoley/ClawManager/blob/main/docs/patches/openclaw-154735-current-main-integration.md) + [upstream note](https://github.com/openclaw/openclaw/pull/154735#issuecomment-6029188080). Current main already owns the shared formatters, so the conflict can be reduced to wiring the two missing operational warnings into `buildStatusHealthRows()` and refreshing focused tests. |
| [#154702](https://github.com/openclaw/openclaw/pull/154702) | Open — **contributed current-main rebase plan** | [hippoley rebase artifact](https://github.com/hippoley/ClawManager/blob/main/docs/patches/openclaw-154702-current-main-rebase.md) + [upstream note](https://github.com/openclaw/openclaw/pull/154702#issuecomment-6029221063). Current main still lacks the operation-level credential timeout; conflicts are primarily Google provider source-layout drift, so the proven fix can be replayed narrowly. |
| [#154728](https://github.com/openclaw/openclaw/pull/154728) | Open — **contributed evidence plan** | [Gateway behavior proof plan](https://github.com/hippoley/ClawManager/blob/main/docs/evidence/openclaw-154728-gateway-proof-plan.md) + [upstream note](https://github.com/openclaw/openclaw/pull/154728#issuecomment-6029090885). The proposal uses the real `agent.wait` handler and production agent-job owners to prove completion survives later queue / gateway-draining observations, including measured one-worker timing. |
| [#154829](https://github.com/openclaw/openclaw/pull/154829) | Open — **contributed evidence + patch proposal** | [Integration note](https://github.com/openclaw/openclaw/pull/154829#issuecomment-6028995347), [exact patch proposal](https://github.com/hippoley/ClawManager/blob/main/docs/patches/openclaw-154829-current-main.patch), and [upstream follow-up](https://github.com/openclaw/openclaw/pull/154829#issuecomment-6029068588). Current-main analysis shows the runner already owns sourceEnv in coreCtx; the fix can stay at the collector/registration boundary. |
| [#154837](https://github.com/openclaw/openclaw/pull/154837) | Open | Maintainer review / landing decision; avoid redundant changes. |

## Public contribution artifacts

- **OpenClaw #154829** — [current-main integration review note](https://github.com/openclaw/openclaw/pull/154829#issuecomment-6028995347) plus [exact current-main patch proposal](https://github.com/hippoley/ClawManager/blob/main/docs/patches/openclaw-154829-current-main.patch) and [upstream follow-up](https://github.com/openclaw/openclaw/pull/154829#issuecomment-6029068588), authored from `hippoley`. The proposal narrows the fix to explicit env ownership at the disk-space health-check boundary without touching the split runner or ambient process env.
- **OpenClaw #166364** — [compatibility-proof scoping note](https://github.com/openclaw/openclaw/pull/166364#issuecomment-6029001497), [compatibility evidence contract](https://github.com/hippoley/ClawManager/blob/main/docs/evidence/openclaw-166364-compatibility-evidence-contract.md), and [upstream handoff](https://github.com/openclaw/openclaw/pull/166364#issuecomment-6029408639), authored from `hippoley`. The subsequent ClawSweeper revision explicitly concluded that the previous migration-proof blocker is unsupported by the diff; the evidence contract turns that conclusion into a concrete base-vs-candidate equivalence check.
- **OpenClaw #154728** — [Gateway behavior proof plan](https://github.com/hippoley/ClawManager/blob/main/docs/evidence/openclaw-154728-gateway-proof-plan.md), [exact regression](https://github.com/hippoley/ClawManager/blob/main/docs/evidence/openclaw-154728-exact-gateway-regression.md), and [upstream evidence note](https://github.com/openclaw/openclaw/pull/154728#issuecomment-6029157467), authored from `hippoley`. The contribution now includes code-level regression text for the real Gateway `agent.wait` path.
- **OpenClaw #166361** — [CI-cost blocker resolution note](https://github.com/openclaw/openclaw/pull/166361#issuecomment-6029161635), authored from `hippoley`. Extracted already-available CI seconds and converted them into exact test-cost wording without mislabeling shared-shard duration as per-file timing. The author immediately triggered a fresh ClawSweeper review; **Revision 6 cleared the blocker**, and the PR was then **merged on 2026-10-07**.
- **OpenClaw #154735** — [current-main integration patch](https://github.com/hippoley/ClawManager/blob/main/docs/patches/openclaw-154735-current-main-integration.md), [exact deep-status regression](https://github.com/hippoley/ClawManager/blob/main/docs/evidence/openclaw-154735-exact-deep-status-regression.md), and [upstream handoff](https://github.com/openclaw/openclaw/pull/154735#issuecomment-6029373784), authored from `hippoley`. Reconciled the stale branch with current main by reusing the formatters that have already landed and reducing the remaining work to status-row wiring plus report-boundary regression.
- **OpenClaw #154702** — [current-main rebase artifact](https://github.com/hippoley/ClawManager/blob/main/docs/patches/openclaw-154702-current-main-rebase.md), [exact current-main patch](https://github.com/hippoley/ClawManager/blob/main/docs/patches/openclaw-154702-exact-current-main.patch.md), and [upstream handoff](https://github.com/openclaw/openclaw/pull/154702#issuecomment-6029622517), authored from `hippoley`. Verified that the credential-timeout defect still exists on current main, narrowed the conflict to source-layout drift, and translated the proven candidate into a directly applicable current-main production patch while preserving omitted-timeout compatibility.

## What counts as a durable GitHub contribution

The goal is not to accumulate references to other people's PRs. The durable target is to leave a public `hippoley` artifact in the upstream chain:

1. a focused PR with a reproducible bug and bounded patch;
2. a runtime or compatibility proof attached to an upstream PR;
3. a review that finds a concrete defect or verifies a hard blocker;
4. an issue with a minimal reproduction and maintainer-usable evidence;
5. a benchmark or timing artifact that resolves an explicit review request.

When an item is merged, the record should be moved out of the active table immediately. If upstream incorporates or supersedes the work, the record should preserve that relationship without relabeling it as a merged `hippoley` PR.

## Near-term objective

Convert at least one active OpenClaw blocker into a public `hippoley` contribution artifact, then link that artifact here with exact provenance.

- **OpenClaw #166371** — [real Claude CLI proof protocol](https://github.com/hippoley/ClawManager/blob/main/docs/evidence/openclaw-166371-real-claude-cli-proof-protocol.md) and [upstream handoff](https://github.com/openclaw/openclaw/pull/166371#issuecomment-6029755805), authored from `hippoley`. Converted the remaining after-fix runtime-proof blocker into a branch-specific Claude CLI transcript/delivery contract without fabricating evidence or conflating the overlapping #153032 patch.
- **OpenClaw #166335** — current target discovered during follow-up scanning. Latest completed review had one remaining native-cleanup validation blocker, but a newer head is already under fresh ClawSweeper review. No `hippoley` upstream comment was added; avoid racing the author's update.
- **OpenClaw #166399** — [real Gateway policy-refusal proof contract](https://github.com/hippoley/ClawManager/blob/main/docs/evidence/openclaw-166399-real-gateway-policy-refusal-proof.md) and [upstream handoff](https://github.com/openclaw/openclaw/pull/166399#issuecomment-6029847890), authored from `hippoley`. Converted the external-contributor proof blocker into a concrete real-Gateway continuity/history contract: safe replacement reply, retained internal session identity, allowed follow-up success, and exclusion of blocked/private policy text from projected history.
- **OpenClaw #155320** — [current-head test-cost handoff](https://github.com/hippoley/ClawManager/blob/main/docs/evidence/openclaw-155320-current-head-test-cost-handoff.md) and [upstream handoff](https://github.com/openclaw/openclaw/pull/155320#issuecomment-6029921245), authored from `hippoley`. Verified exact-head CI success and green repository gate, explicitly refused to mislabel shared workflow duration as per-file timing, and reduced the last blocker to one focused consolidated-head wall-time measurement.

## 2026-10-07 self-audit / refresh

- **OpenClaw #166335** — fresh Revision 6 now reports **Ready for maintainer review**, proof 5/6, patch 5/6, and **Before merge: None**. The successful non-container native Doctor acceptance resolves the earlier cleanup-proof blocker. No `hippoley` upstream comment was needed; the correct action was to avoid racing the author's evidence.
- **OpenClaw #154829 artifact self-audit** — upgraded `docs/patches/openclaw-154829-current-main.patch` from a production-only exact diff plus prose regression guidance into a full v2 patch that includes the collector unit regression and the real mixed selected-check integration hunk. Verified current helper signatures, `tryReadDiskSpace` spy shape, and existing SQLite-preservation fixture before updating.
- **OpenClaw #154735 artifact self-audit** — hardened `docs/evidence/openclaw-154735-exact-deep-status-regression.md` against current `HealthSummary` protocol schema. Replaced direct fixture mutation with typed object-spread construction and verified the exact protocol fields and formatter output used by the expected rows.

- **OpenClaw #148397** — [thinking-provenance compatibility contract](https://github.com/hippoley/ClawManager/blob/main/docs/patches/openclaw-148397-thinking-provenance-contract.md), authored from `hippoley`. Re-checked current head `7a88f9441f37d076b6b6c8dbf4529e812748aab7` and confirmed the stale P1 remains: declared Qwen chat-template models can still send mapped `high` on known-unselected turns. Root cause is loss of the existing `thinkingExplicit` provenance before the provider boundary. The contract was self-audited down to a minimal run-scoped path that avoids persisted state, AgentState, AgentLoopConfig, and llm-core API expansion, and defines three-state rollout semantics for reply, agent-exec, and cron owners. Fresh ClawSweeper review is currently in progress; no upstream comment was added to avoid racing the new verdict.