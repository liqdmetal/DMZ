# Wrap acceptance citation triage

**Date:** 2026-11-10
**Tree:** `chore/wrap-recovery` at `1ae7d74` (`674836f` plus the recovered Monero PR `e15e442`); `wm`/`wv`
  are still not the cited trees. Recovery record: `research/wrapper/evidence/RECOVERY_2026-11-10.md`.
**Source:** `docs/WRAP_ACCEPTANCE_MATRIX.md` — all 74 rows

This is a citation and implementation audit of the merged tree, not an acceptance run. It resolves every prefixed
`main:`, `wc:`, `zfee:`, `wm:`, `wz:` and `wv:` citation, checks the cited file and line range, and then asks whether
the row's **required mitigation** is present. A row is not `implemented` merely because the code it cites exists:
many rows cite the code that demonstrates the gap.

## Method

- `main:` paths were mapped to their merged locations (`vault.bas` → `contracts/vault.bas`, `main.go` →
  `tools/vaultsigner/main.go`, `wrap.js` → `app/wrap.js`, and so on).
- `wc:` and `zfee:` landed on main and were read there. `wz:` was read under `tools/vaultsigner` and
  `tools/zcashtx`.
- `wm:` and `wv:` still have no matching tree. The recovered `feat/vaultsigner-monero-retries` branch implements
  `tools/vaultsigner/monerovault`, not the cited `xmrvault`/`monero.go` paths, and `wv` has no ref. Their 43 and 27
  cite tokens remain unresolved; a row that needs them cannot be `implemented` against this tree.
- Every prefixed cited line range was checked against the current file. `WT-45`'s old `zfee:wrapkit.js:62-64`
  reference is past the current file end and has been replaced with the current config/formatting code. `WT-58`'s
  document ranges were refreshed to the passages that now carry the contradictions; the row moves from `stale` to
  `partial` because the underlying document rewrite and `doccites` test are still absent.
- Bare follow-on cites (`main.go:864`, `backend.go:670-713`) inherit the file from the preceding prefixed cite. They
  were read in context where they decide a row's state, but are not separate rows in the table.
- The `Merged code` column lists the files whose citations resolve. It does **not** mean those files implement the
  row; the `Why` column says what is still missing.

## State definitions

| State | Meaning |
|---|---|
| `implemented` | Every part of the required mitigation is present in the merged tree at the cited code. |
| `partial` | At least one required piece is present, and at least one is absent — usually an unmerged `wm`/`wv` path, or a concrete missing change on main. |
| `missing` | No required mitigation is merged. The citations either do not resolve or describe the gap itself. |
| `stale` | The row's citations are themselves outdated and must be rewritten before the row can be assessed. |

## Counts

| State | Critical | High | Medium | Low | Total |
|---|---:|---:|---:|---:|---:|
| implemented | 0 | 0 | 0 | 0 | **0** |
| partial | 15 | 17 | 11 | 4 | **47** |
| missing | 4 | 11 | 9 | 3 | **27** |
| stale | 0 | 0 | 0 | 0 | **0** |
| **total** | **19** | **28** | **20** | **7** | **74** |

## Findings

- **No row is fully implemented against the merged tree.** The closest are rows with correct merged code but an
  unmerged backend or a test-only gap; they remain `partial`.
- **Every critical row is short of its required mitigation.** Fifteen are `partial`; `WT-08`, `WT-11`, `WT-13` and
  `WT-14` are `missing`.
- **The largest single cause is the missing v4/Monero trees.** A Monero backend branch was recovered and merged,
  but it is not the cited `wm` code; `wv` remains absent. The Monero release path, v4 contract and v4 daemon cannot
  be verified or counted from this tree.
- **`wz:`, `wc:` and `zfee:` resolve.** The Zcash backend, coordination nodes and the Deposit & withdraw hub are
  present under their merged paths; their rows are still partial where a required change was never made.
- **Citation drift is real but narrower than the branch gap.** WT-45's out-of-range `wrapkit.js` cite and WT-58's
  document ranges are now refreshed. WT-58 is `partial`, not `stale`: its citations resolve, but the doc rewrite,
  gas reconciliation and `doccites` test are still missing. The matrix's other `main:` cites still need their own
  regeneration pass.
- **The app-side WT-17 and WT-69 fixes are present in this tree.** Both rows remain `partial`, not `implemented`,
  because WT-17 still needs its DVM/uitest checks and WT-69 still needs simulator e2e; the v4 contract/daemon cites
  are also still unavailable here.

## Recovery rerun (2026-11-10)

`chore/wrap-recovery` merges the recovered `e15e442` Monero backend on top of the work-board commit. It is
green (`go test ./vaultsigner/... ./zcashtx/...`; JS 148/148), but it does not resolve the audit:

- `wm:` has 43 tokens; none points at a file that exists in the recovered branch (`monerovault`, not `xmrvault`).
- `wv:` has 27 tokens; no `vault.bas` v4, `tests/vault_v4_test.go` or `wv:main.go` exists. The app-side v4 read is
  present, but the cited contract and daemon are not.
- The row states are therefore unchanged. The full search record is `research/wrapper/evidence/RECOVERY_2026-11-10.md`.

## Rows

| ID | Severity | State | Merged code that resolves | Missing / stale | Why |
|---|---|---|---|---|---|
| WT-01 | critical | **partial** | `contracts/vault.bas`, `tools/vaultsigner/main.go`, `tools/vaultsigner/zcashvault/build.go`, `app/move.js` | wm:xmrvault/signer.go:568-625 | v3 contract/daemon and Zcash exist; v4 `nfee` rule and Monero credit removal do not. |
| WT-02 | critical | **partial** | `tools/vaultsigner/main.go`, `contracts/vault.bas` | wv:vault.bas:463, wv:vault.bas:500-503 | v3 freeze and daemon checks exist; v4 Reclaim/mark window does not. |
| WT-03 | critical | **partial** | `tools/vaultsigner/main.go`, `contracts/vault.bas` | wv:vault.bas:463, wv:main.go:562-566 | main daemon/contract exist; only Released has a fixed fee; v4 reset does not. |
| WT-04 | critical | **partial** | `tools/vaultsigner/zcashvault/build.go` | wm:xmrvault/release.go:47-55, wv:vault.bas:489-494 | Zcash release protection exists; Monero cancel path and v4 comment correction do not. |
| WT-05 | critical | **partial** | `tools/vaultsigner/zcashvault/backend.go`, `tools/vaultsigner/coord/node.go` | wm:xmrvault/release.go:57-60, wm:xmrvault/ceremony.go:72-74 | Zcash rule exists; coordination keeps the latest version only; Monero missing. |
| WT-06 | critical | **partial** | `tools/vaultsigner/main.go`, `tools/vaultsigner/zcashvault/backend.go` | wm:xmrvault/release.go:266-313 | Main retry and Zcash lookup exist; Monero lookup and verify-depth change do not. |
| WT-07 | critical | **partial** | `tools/vaultsigner/coord/node.go`, `tools/vaultsigner/zcashvault/backend.go`, `tools/vaultsigner/main.go` | wm:xmrvault/release.go:72-81 | Zcash/coordination/main pieces exist; Monero ledger rule and v4 Paying commit do not. |
| WT-08 | critical | **missing** | `research/wrapper/RUNBOOKS.md` | wm:xmrvault/release.go:490-505 | Only a runbook that copies files; no restore/export rule and no Monero code. |
| WT-09 | critical | **partial** | `tools/vaultsigner/zcashvault/backend.go` | wm:xmrvault/signer.go:399-403 | Zcash change address is handled; `ValidDest` does not refuse vault scripts; Monero missing. |
| WT-10 | critical | **partial** | `tools/vaultsigner/zcashvault/backend.go` | wm:xmrvault/release.go:601-666 | Zcash signature rebuild exists; Monero `checkSet` proof is missing. |
| WT-11 | critical | **missing** | `tools/vaultsigner/main.go` | — | NONE in the matrix; the cited `LiveAt` state is the gap, not an anchor. |
| WT-12 | critical | **partial** | `tools/vaultsigner/main.go`, `contracts/vault.bas`, `tools/vaultsigner/zcashvault/backend.go` | wm:xmrvault/signer.go:545-547 | Zcash `da_` mismatch guard exists; main still accepts a second quorum; Monero missing. |
| WT-13 | critical | **missing** | — | wm:xmrvault/signer.go:436 | No public-key dedupe or unique key-image reserve in the merged tree. |
| WT-14 | critical | **missing** | `tools/vaultsigner/zcashvault/backend.go` | — | NONE; the cited Zcash output loop is the gap. |
| WT-15 | critical | **partial** | `tools/vaultsigner/main.go`, `contracts/vault.bas`, `tools/vaultsigner/zcashvault/backend.go` | wv:vault.bas:454-467, wm:xmrvault/signer.go:587-595 | Main `reserve >= sup - inflight` exists; marked/verified-supply rule and Monero reserve do not. |
| WT-16 | critical | **partial** | `tools/vaultsigner/zcashvault/descriptor.go`, `contracts/vault.bas`, `contracts/README.md`, `app/wrap.js` | wm:xmrvault/ceremony.go:72-74, wv:vault.bas:260, wv:vault.bas:580 | Zcash enforces 2<=k<n<=15; main contract allows n=20; operator ruling/Monero missing. |
| WT-17 | critical | **partial** | `app/move.js`, `app/vault.js`, `app/wrap.js` | wv:vault.bas:285-292, wv:app/vault.js:32-33 | v3 raw and v4 `ADDRESS_STRING` reads now both work; the Burn probe checks its own record. DVM and uitest are still missing. |
| WT-21 | critical | **partial** | `tools/vaultsigner/main.go`, `tools/vaultsigner/zcashvault/backend.go` | wm:xmrvault/signer.go:311-314 | Main re-verifies at vote time; no second independent node; Monero witness check missing. |
| WT-66 | critical | **partial** | `contracts/vault.bas`, `tools/vaultsigner/main.go`, `tools/vaultsigner/zcashvault/build.go` | wv:vault.bas:454-467, wm:xmrvault/release.go:685-745 | Inflight counting exists; the verified-supply gate is absent. |
| WT-18 | high | **partial** | `contracts/vault.bas`, `tools/vaultsigner/coordwire.go`, `tools/vaultsigner/zcashvault/backend.go` | — | Coordination warning and Zcash xpubs exist; daemon `isg_` check and custody epoch do not. |
| WT-19 | high | **partial** | `deploy/VAULT_PARAMETERS.md`, `contracts/vault.bas` | wv:vault.bas:578, wv:vault.bas:595-597 | Main bond arithmetic/docs exist; v4 Retire and slash-to-holders rule do not. |
| WT-20 | high | **partial** | `contracts/vault.bas` | — | A `winmax` guard exists, but Seal ignores it; the required Seal/limit rule is absent. |
| WT-22 | high | **partial** | `tools/vaultsigner/networks.go`, `tools/vaultsigner/zcashvault/backend.go` | wm:xmrvault/signer.go:127-128, wm:xmrvault/signer.go:118-119 | Zcash finality and main blocks exist; daemon/backend floors still differ; Monero missing. |
| WT-23 | high | **missing** | `tools/vaultsigner/main.go` | — | Pays straight after the mark; the stable-height re-read is absent. |
| WT-24 | high | **missing** | `tools/vaultsigner/main.go`, `tools/vaultsigner/zcash_backend.go`, `tools/vaultsigner/zcashvault/build.go` | — | The loop still uses `log.Fatal`/whole-pass failure; per-index isolation is absent. |
| WT-25 | high | **partial** | `contracts/vault.bas` | wv:vault.bas:292 | Some vote/cap checks exist; the own-vote/window/budget/regfee changes do not. |
| WT-26 | high | **missing** | `tools/vaultsigner/main.go`, `contracts/vault.bas` | — | Current code casts Halt(0); the required never-Halt(0)/consecutive-pass rule is absent. |
| WT-27 | high | **partial** | `tools/vaultsigner/main.go` | — | Halted skip for unmarked burns exists; the Lookup/Pay split and co-signer recheck do not. |
| WT-28 | high | **partial** | `tools/vaultsigner/coord/identity.go`, `tools/vaultsigner/coordwire.go` | wm:xmrvault/release.go:107-111 | Ed25519 identity proof exists; on-chain DERO-signature binding and active-set filtering do not. |
| WT-29 | high | **partial** | `tools/vaultsigner/zcashvault/descriptor.go` | — | Zcash vectors and branch-id check exist; the upgrade signing gate is absent. |
| WT-30 | high | **missing** | — | wm:xmrvault/signer.go:214-232, wm:monero.go:100-101 | No Monero backend in the merged tree; the node/witness/ceremony work is absent. |
| WT-31 | high | **partial** | `contracts/vault.bas`, `tools/vaultsigner/zcashvault/backend.go` | wm:xmrvault/signer.go:399-403 | Zcash real vout exists; the shared written rule/golden vectors are absent; Monero missing. |
| WT-32 | high | **partial** | `contracts/vault.bas`, `app/wrap.js`, `tools/vaultsigner/main.go`, `tools/vaultsigner/zcashvault/backend.go` | wm:xmrvault/signer.go:551-566 | Zcash transparent checks exist; app length-only check and veto-mark rule remain; Monero missing. |
| WT-33 | high | **partial** | `tools/vaultsigner/zcashvault/descriptor.go` | wm:xmrvault/release.go:82-84 | Zcash binds the ledger to the vault; Monero vault binding is missing. |
| WT-34 | high | **missing** | `app/wrap.js`, `tools/vaultsigner/zcash_backend.go` | wm:monerorpc/address.go:254 | The app displays a returned `da_`; it does not derive or cross-check it. |
| WT-35 | high | **partial** | `app/move.js`, `contracts/vault.bas`, `app/vault.js` | wv:vault.bas:500-503 | Hub still scans only the newest 60 and no daemon calls Reclaim; the app now lists v3/v4 withdrawals and sends Reclaim at a fixed fee. |
| WT-37 | high | **missing** | `app/wrap.js`, `app/move.js` | — | The screen reads balance; `mint_`/`mj_`/`ma_`/`mh_`/`rf_` status is absent. |
| WT-38 | high | **partial** | `app/wrap.js` | wv:app/vault.js:32-33 | Kind/version checks exist; `templates.vault` code-hash matching and legacy behavior do not. |
| WT-39 | high | **missing** | `app/wrapui.js` | — | One `sym` is used on both sides; the XMR/wXMR naming rule is absent. |
| WT-40 | high | **partial** | `contracts/vault.bas`, `app/wrap.js`, `app/move.js`, `tools/vaultsigner/zcashvault/descriptor.go` | — | Zcash custody copy exists; no tick check, Monero publication or daemon SetViewKey refusal. |
| WT-41 | high | **partial** | `contracts/vault.bas` | wv:vault.bas:62 | SetViewKey/verifydeploy pieces exist; v4 `SetFKey`/custody hash and Seal/Register gating do not. |
| WT-47 | high | **missing** | `tools/simvault/main.go`, `docs/SIGNER_RUNBOOK.md` | — | No `vaultctl`; simvault hard-codes wallets; no compromise runbook/sweep. |
| WT-67 | high | **missing** | `tools/vaultsigner/main.go` | wm:monero.go:100-101 | No loopback/login enforcement in vaultsigner; Monero checks are unmerged. |
| WT-68 | high | **missing** | `tools/vaultsigner/main.go`, `tools/vaultsigner/zcashvault/backend.go` | wm:monero.go:268-279, wm:monero.go:128-133 | One node is accepted; no second independent node or stable-height cross-check. |
| WT-69 | high | **partial** | `app/app.js`, `app/wrap.js`, `contracts/vault.bas` | wv:vault.bas:501-503, wv:vault.bas:283, wv:vault.bas:292, wv:main.go:873, wv:vault.bas:462-463 | Reclaim and Register now use a fixed capped fee and skip the height-0 estimator; simulator e2e is still missing. |
| WT-70 | high | **missing** | `tools/vaultsigner/zcashvault/build.go` | wm:xmrvault/release.go:117-124, wm:monero.go:159-165, wv:vault.bas:580 | No sweep code; the runbook describes a process the tree does not implement. |
| WT-71 | high | **partial** | `contracts/vault.bas`, `tools/vaultsigner/main.go`, `tools/vaultsigner/zcashvault/build.go` | wm:xmrvault/release.go:180-188 | Contract `fu_` check exists; the daemon-side adopt/refuse path is incomplete. |
| WT-36 | medium | **partial** | `app/wrap.js` | — | A confirmation sheet exists; reload-before-open, absolute heights and fingerprint are absent. |
| WT-42 | medium | **partial** | `contracts/vault.bas`, `tools/vaultsigner/zcashvault/backend.go` | wm:xmrvault/signer.go:459 | Per-output accounting exists; the per-(txid,j) aggregation/disclosure is absent. |
| WT-43 | medium | **partial** | `tools/vaultsigner/main.go`, `app/move.js` | — | Minting exists in `Deposits()` order; oldest-first/`mw_`/max-deposit gating is absent. |
| WT-44 | medium | **missing** | `tools/vaultsigner/main.go` | — | Refunded is a manual k-vote with no tool; deterministic Zcash refund/ledger is absent. |
| WT-45 | medium | **partial** | `app/config.js`, `app/wrapui.js`, `tools/vaultsigner/zcash_backend.go` | wm:xmrvault/signer.go:219-221 | Config sets XMR `dec` 12 / ZEC `dec` 8; `wrapui.js:10` only bounds 0–18 and no app code checks 12/8. `Decimals` is unread and no contract/deployment rule exists. |
| WT-46 | medium | **partial** | `app/move.js` | wv:vault.bas:267 | The v3 table exists; v4 arithmetic, verifydeploy v4 and app version split are absent. |
| WT-48 | medium | **missing** | `tools/vaultsigner/main.go`, `tools/vaultsigner/coordwire.go` | — | Only logs; loopback `/health`, `/metrics` and `vaultwatch` are absent. |
| WT-49 | medium | **partial** | `tools/vaultsigner/main.go` | — | Mainnet separation exists; the `DERO.GetInfo` network comparison is absent. |
| WT-50 | medium | **partial** | `contracts/vault.bas`, `tools/vaultsigner/zcashvault/build.go` | wm:xmrvault/release.go:160-168 | v3 `fu_` string binding exists; v4 raw `ftx` and durable txid->j map are absent. |
| WT-51 | medium | **missing** | `contracts/vault.bas` | — | Only the burn value is compared; `wopen` tracking/delay is absent. |
| WT-52 | medium | **missing** | `contracts/vault.bas` | — | Epoch scopes only; no expiry bucket or vote withdrawal. |
| WT-53 | medium | **missing** | `contracts/vault.bas` | wm:xmrvault/release.go:247-264 | No change splitting or throughput cap; Monero signing path missing. |
| WT-54 | medium | **partial** | `tools/vaultsigner/zcashvault/backend.go` | wm:xmrvault/signer.go:225-227 | Zcash node agreement exists; no unknown-version fork gate; Monero missing. |
| WT-55 | medium | **partial** | `tools/zcashtx` | — | `zcashtx` has transparent v6/ZIP229; shielded v6 is refused and no app warning exists. |
| WT-56 | medium | **partial** | `tools/vaultsigner/zcashvault/backend.go` | wm:xmrvault/release.go:626-644 | The ordering is right in the merged code, but the Monero half is unmerged and the ordering test is wrong. |
| WT-57 | medium | **missing** | — | wm:xmrvault/signer.go:533-535 | No Monero key-image coverage check in Health. |
| WT-58 | medium | **partial** | `research/wrapper/RUNBOOKS.md`, `research/wrapper/README.md`, `deploy/VAULT_PARAMETERS.md`, `docs/HANDOVER.md`, `contracts/README.md` | — | Citations now resolve to the current contradictory passages, but the v4 document rewrite, gas reconciliation, `doccites` test and matrix-wide cite pass are still absent. |
| WT-59 | medium | **missing** | — | wv:tests/vault_v4_test.go:42 | `tests/vault_v4_test.go` and `TestVaultV4InstallGas` are absent. |
| WT-72 | medium | **missing** | `tools/vaultsigner/zcashvault/build.go`, `contracts/vault.bas` | wm:xmrvault/release.go:690-692 | No consolidation round or fee accounting for it. |
| WT-73 | medium | **missing** | `tools/vaultsigner/zcashvault/seed.go`, `tools/vaultsigner/zcash_backend.go` | — | Seed is plaintext hex; no encryption/OS keystore or mainnet refusal. |
| WT-60 | low | **missing** | — | — | NONE; only Burn reads `ASSETVALUE`; the guard is absent. |
| WT-61 | low | **partial** | `tools/vaultsigner/zcashvault/backend.go`, `tools/vaultsigner/main.go` | wm:xmrvault/signer.go:420 | Zcash height guard exists; Monero missing and the daemon does not skip. |
| WT-62 | low | **partial** | `tools/vaultsigner/zcashvault/backend.go` | wm:xmrvault/signer.go:527 | Zcash guards exist; Monero guards are missing. |
| WT-63 | low | **missing** | — | — | The v4 Burn comment/detection is absent. |
| WT-64 | low | **missing** | `app/wrap.js` | — | `Claims outstanding` is one number; the breakdown/invariant is absent. |
| WT-65 | low | **partial** | `app/move.js` | — | `move.js` has raw/string matching; it does not yet carry all WT-17/34/38 guards. |
| WT-74 | low | **partial** | `tools/vaultsigner/coordwire.go`, `contracts/vault.bas` | wv:vault.bas:199, wm:monero.go:159-165 | The 64-seat cap exists; paged reads of all n seats are absent. |

## What this does not change

The board's `open` / `in-progress` / `blocked` / `done` column records acceptance work. These states are a separate,
stricter code audit. No row becomes `done` from this document; the WT-17 and WT-69 passing unit logs are app-side
evidence, and both rows still need the DVM/uitest or simulator e2e the matrix requires.
