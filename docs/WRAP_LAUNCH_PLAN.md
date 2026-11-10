# DMZ wrap launch plan: production grade, then mainnet

*2026-11-09. A plan, not evidence. Every state claimed here cites the tree or a document; nothing is marked done
because it is expected to be done. The acceptance criteria are `docs/WRAP_ACCEPTANCE_MATRIX.md` (74 rows: 19
critical, 28 high, 20 medium, 7 low). The product decision is the operator's ruling of 2026-09-15: the atomic swaps
are removed, and mainnet waits until wXMR and wZEC both work end to end through the federation.*

## 1. What is being launched

- **wXMR and wZEC**, each a 1:1 claim on XMR or ZEC held by a k-of-n federation of named third-party signers,
  never called trustless.
- **DMZ markets** on DERO, with the wrapped assets as quote assets.
- **Both assets on mainnet.** The operator's rule is "wait for both wrapped"; a wZEC-only launch is not the
  approved product.
- **Not in scope:** the removed atomic swaps, any bridge claim, the read-only Solana gallery (independent,
  browser-only, no custody), and any fixed-rate "1:1 market" until the operator defines it (YB-25).

"Custody" is the correct description. A colluding quorum can take the reserve; the contract makes the DERO side
auditable, not the foreign side safe.

## 2. Current position

| Area | State in this workspace | Evidence | What launch still needs |
|---|---|---|---|
| Contracts | `contracts/vault.bas` is v3. Vault v4 is referenced by the matrix as `wv:` and is **not in this tree**. The DVM suite last recorded 175 runs, 0 failures on 2026-09-12. | `contracts/vault.bas:1-9`; `docs/MAINNET_TODO.md` Y3-3, Y5-1 | v4 merged, reviewed, frozen; every confirmed finding fixed with a regression test; external review closed |
| Monero backend | `tools/vaultsigner/monerovault` reads deposits, balance and addresses and pays through a **single-key** wallet. A `multisig_txset` is refused with `ErrNeedsSignatures`. | Branch `feat/vaultsigner-monero-retries` (`e15e442`): `monerovault/backend.go` `Release`; `monerovault/client.go` `ErrNeedsSignatures` | A Monero multisig signing coordinator: txset distribution, per-signer verification, `sign_multisig`, `submit_multisig`, cancel/replacement rules. This is the critical-path item |
| Monero production work | The matrix cites `wm:` (`wrap-monero`) files such as `xmrvault/release.go`, `xmrvault/ceremony.go` and `monerorpc/`. **No such branch or worktree exists here**, and no remote branch contains them. The matrix says those worktrees were mid-edit with uncommitted changes. | `docs/WRAP_ACCEPTANCE_MATRIX.md` header and WT-04, WT-10, WT-30, WT-71 | Recover the worktrees, or rebuild from `docs/HANDOVER.md` §5.3 and the matrix's required tests. Do this before any schedule commitment |
| Zcash backend | Transparent P2SH k-of-n is built and ran a live 2-of-3 wrap and unwrap on zebrad regtest. The live offline-signer withdrawal is **not** proven: Y3-1 records it only against an in-memory validating chain. | `docs/HANDOVER.md` §5.4; `research/regtest/vault-zcash-e2e.log`; Y3-1 | A live offline-signer run; public-testnet pilot with real signers; external review; reserve consolidation; sub-`mindep` refunds; encrypted seed; node cross-checks |
| Signer daemon and coordination | HPKE + AES-GCM + Ed25519 coordination landed; retries keep a signer alive through node restarts. | `tools/vaultsigner/coord/`; branch `feat/vaultsigner-monero-retries` (`e15e442`): `tools/vaultsigner/node.go`; Y3-4 | Health/metrics endpoint, alerting, sealed-instance lease, mainnet wallet-login and loopback enforcement, full active-set reads |
| Vault deployment | No mainnet install path. `simdeploy` has no vault path, `simvault` has no `-network`, and the mainnet bundle excludes a vault. | `deploy/README.md` "Vault: not on mainnet"; YB-13 | A rehearsed install/seal/verify path; `deploy/federation.md`; parameters, bonds and caps |
| App | The wrap screens and Deposit & Withdraw hub exist. The matrix still lists one critical app defect (WT-17: a successful Burn or Register reported as "did not execute", so users burn again) and several high ones: height-gated estimates (WT-69), unverified deposit addresses (WT-34), status claims the contract does not make (WT-37), unpinned vault code (WT-38), and wrapped-unit labelling (WT-39). | `docs/WRAP_ACCEPTANCE_MATRIX.md` WT-17, 34, 37-40, 69 | Every app row closed with its UI test; a mainnet-shaped pass that expects a vault, not the old no-vault build (YB-19) |
| Build and test | Focused Go tests and `go vet` for `vaultsigner`, `zcashtx` and `simrpc` pass on the signer commit. `tools/check.sh` cannot run here: `vendor/` is absent and Node is not installed. Local Go is **1.27.3**, while the gate's evidence pins **go1.26.8**. The clone is shallow (`grafted`). | `tools/check.sh`; `go version`; `git log` | A full, reproducible toolchain; full `check.sh` green; a complete clone for release provenance |
| Submission | The work is pushed to `liqdmetal/DMZ` only. The account cannot push to `NoahMoney/DMZ`. | `gh api repos/NoahMoney/DMZ` permissions; `fork` remote | Canonical-repo write access and a release-signing decision (YB-20) before any artifact is called a release |

**The matrix is not a status board.** Its "Current protection" column describes code, not accepted risk. A row is
accepted only when its required mitigation and required tests exist.

## 3. Launch gates

No gate can be waived by schedule pressure. Each is a go/no-go with evidence attached.

| Gate | Requirement | Evidence |
|---|---|---|
| **G0 Recovery** | `wrap-monero`, `wrap-zcash` and `wrap-vault` are recovered or rebuilt, committed, and build from a full clone. | Branch commits; `go build ./...`; no uncommitted worktree |
| **G1 Baseline** | Pinned toolchain (go1.26.8), `vendor/derohe`, pinned Node, and `tools/check.sh` green from a clean checkout. | `check.sh` log; `go version`; `verifyvendor` |
| **G2 Freeze** | Vault v4 and both backends frozen at named commits; contract hashes published; external review package sent. | `docs/audit/`; hashes; reviewer engagement letter |
| **G3 Safety** | Zero open critical rows by the end of P2; zero open critical or high rows by the end of P4. Each closed row names its test. | Matrix with status; test log per row |
| **G4 Federation E2E** | wXMR and wZEC each run deposit → mint → burn → release → Released on a public testnet with 2-of-3, one signer offline, and a restart mid-payment. | Logs under `research/wrapper/evidence/`, no machine paths |
| **G5 Operations** | Monitoring, alerting, incident drills, key ceremony, backup/restore and `vaultctl` rehearsed on testnet. | Drill logs; runbook review |
| **G6 External review** | Contracts, vault, daemon and both backends reviewed; every finding fixed or accepted in writing by the operator. | Review report; remediation table |
| **G7 Pilot** | At least one week (preferably two) of real testnet deposits and withdrawals at low caps with named signers, no unexplained halt, and proof of reserves checked independently. | `research/wrapper/evidence/`; daily status |
| **G8 Mainnet rehearsal** | The mainnet procedure runs end to end on testnet with mainnet flags, fresh wallets, `verifydeploy` clean, app on a ring-16 INDEX, treasury dry run. | Timed rehearsal log; cost table |
| **G9 Launch** | G0-G8 plus operator sign-off, funded wallets, published disclosures, and a named on-call rotation. | Signed checklist |

## 4. Phases

Durations are indicative for a small team working in parallel. The external review and signer recruitment are the
calendar-critical items, not the coding.

### P0 — Recover and establish the truth (week 1)

**Objective:** know what code exists before planning around it.

1. Locate the `wrap-monero`, `wrap-zcash` and `wrap-vault` worktrees on the operator's machines. Commit and push
   every one, even mid-edit, on its own branch. If they are lost, say so immediately and rebuild; do not let the
   plan assume they exist.
2. Create one integration branch from `origin/main`. Merge the signer retry commit and the acceptance-matrix work.
   Keep the pool-safety and Solana gallery changes on their own branches.
3. Restore the build baseline: `vendor/derohe`, Go 1.26.8, pinned Node, a full (non-shallow) clone.
4. Run `tools/check.sh` and record every failure. This is the first honest baseline.

**Exit:** G0 and G1.

### P1 — Integrate, freeze, and package for review (weeks 1-3)

**Objective:** one candidate, one contract version, one review package.

1. Merge the three worktrees into the integration branch. Resolve the vault v3/v4 split in one place: the daemon,
   app, verifydeploy, genconfig and docs must all speak the same version.
2. Freeze `vault.bas` v4. Close the contract rows that change its shape first: fee accounting (WT-01), Paying mark
   and Reclaim (WT-02, WT-03), canonical `ftx` (WT-50), split-burn delay (WT-51), vote expiry (WT-52), attestation
   replacement (WT-12), signer-set shape (WT-16), mint window (WT-20), and storage-gas budget (WT-59).
3. Rebuild `docs/audit/CONTRACT_REVIEW_PACKAGE.md` against the frozen hashes and send it. External review has a long
   lead time; start it before the code is perfect, then send a delta.
4. Convert `docs/WRAP_ACCEPTANCE_MATRIX.md` into a tracked work board: one row per WT item, owner, status, test
   name, evidence path. Re-check every `wm:`, `wz:`, `wv:` and `main:` cite against the merged tree (the matrix
   warns they are stale).

**Exit:** G2.

### P2 — Close the critical safety rows (weeks 3-7)

**Objective:** no known way to lose or duplicate funds.

Work the 19 critical rows in dependency order:

- **Money conservation:** WT-01 (network fee on chain), WT-09 (change minted as a deposit), WT-13 (unspendable
  Monero outputs), WT-14 (Zcash coinbase), WT-15 (blind reserve check), WT-66 (DERO keys alone mint an unbacked
  deposit and burn it).
- **Exactly-once and liveness:** WT-02 (a lapsed Paying mark freezes or double-pays), WT-03 (the height-0 estimator
  makes every gated call unreachable), WT-04 (withheld signed payment), WT-05 (equivocating proposer), WT-06 (late
  confirmation), WT-07 (lost state / two daemons), WT-08 (restored Monero wallet reuses nonces).
- **Deposit attribution:** WT-11 (local `LiveAt` loses deposits), WT-12 (attested address not checked).
- **Signing safety:** WT-10 (co-signer signs a different transaction), WT-16 (contradictory k/n rules).
- **Quorum compromise:** WT-21 (a false mint completed by one lazy signer goes unpunished).
- **App path:** WT-17 (a successful Burn or Register reported as failed).

The high rows that are launch-blocking even though they are not critical — height-gated estimates (WT-69),
reorganisation response (WT-22), mark margin (WT-23), pass failure isolation (WT-24), vote spam (WT-25), halts
(WT-26, WT-27), coordination identity (WT-28), Zcash script and fee construction (WT-29) — land in the same phase.

Every row lands with the test the matrix names, and the test is shown failing on the old code where possible.

**Exit:** all critical rows closed; high rows that can lose funds, duplicate a payment or halt the vault closed too
(WT-20 to WT-31, WT-67, WT-68). The remaining high rows — app semantics, disclosures, deployment order, docs and
consolidation — close in P3 and P4, before G3 is assessed at the end of P4.

### P3 — Federation end to end on public testnets (weeks 6-9)

**Objective:** both assets work with the real signing paths, not a single wallet.

1. **Monero:** finish the multisig coordinator. A Monero txset is built by one wallet, verified by every signer
   with `describe_transfer`, signed with `sign_multisig`, and submitted once. The round must bind `(vault, j,
   attempt, proposal hash)`, publish partial signatures, and define the cancel path (WT-04). Run 2-of-3 and 3-of-5
   on stagenet with one signer offline.
2. **Zcash:** run the existing P2SH path on Zcash testnet through `-coord` with at least two nodes per signer.
   Close the reserve-consolidation row (WT-72), the v6 deposit row (WT-55), and the hot-seed row (WT-73).
3. **Cross-chain operations:** fee accounting per chain (WT-01), reorg handling (WT-22), mark margin (WT-23),
   queue throughput (WT-53), sub-`mindep` refunds (WT-44).
4. **Evidence:** every run logs to `research/wrapper/evidence/`, with no machine paths, and records txids for both
   chains.

**Exit:** G4.

### P4 — Make it operable (weeks 8-11)

**Objective:** a signer can run for a year without the operator watching every pass.

1. **Observability:** loopback `/health` and `/metrics`; a `vaultwatch` that alerts on last vote per signer, marks
   near lapse, overdue withdrawals, `na ≤ k+1`, cap/window state, missing view key, halt expiry and fee-wallet
   balance (WT-48).
2. **Control plane:** `vaultctl` with one subcommand per k-vote entrypoint, `-yes` on mainnet, dry run first and
   read-back (WT-47).
3. **Custody operations:** key ceremony runbooks for Monero multisig and Zcash P2SH; encrypted backups; restore
   drills; no file-copy migration of a Monero signer (WT-08, WT-12, WT-70, WT-73).
4. **Node topology:** every signer runs its own loopback node plus one independent agreeing node per chain; no two
   signers share an operator (WT-68). Enforce mainnet wallet login and loopback (WT-67).
5. **Incident response:** vault rows in `INCIDENT_RESPONSE.md`; a compromised-signer drill with the sweep and
   replacement (WT-70); time budgets recorded.

**Exit:** G3 (final assessment) and G5.

### P5 — External review and remediation (weeks 8-14, parallel with P3/P4)

**Objective:** an outside party has tried to break it, and the findings are closed.

Scope: `vault.bas`, `tools/vaultsigner`, `monerorpc`/`monerovault`, `zcashvault`/`zcashtx`, and the app's transaction
and custody surfaces. The review package must include the threat model, the trust statement, every invariant and
its test, known issues, and the open operator decisions. Fix findings on the integration branch; re-run the
affected acceptance rows; send a delta for re-review.

**Exit:** G6.

### P6 — Testnet pilot (weeks 14-16)

**Objective:** real behaviour over time, with real signers.

Install both vaults on the DERO testnet with the final contracts and a ring-16 INDEX. Named signers run their own
hosts and nodes. Caps start low. Users deposit and withdraw both assets daily. Run one planned incident drill and
one unplanned-failure drill. Publish the reserve proof and check it from a machine that is not a signer.

**Exit:** G7.

### P7 — Mainnet rehearsal and launch (weeks 16-19)

**Objective:** execute the written procedure, then execute it for real.

1. Fresh simulator rehearsal of the whole install: contracts, vaults, bonds, `SetLimits`, `Seal`, `SetViewKey`,
   promo, rate, pass with image doc, app DOCs and INDEX. `verifydeploy` clean.
2. Testnet dress rehearsal of `MAINNET_DEPLOYMENT.md` with mainnet flags and fresh wallets; time every step and
   record the DERO cost.
3. Deploy at the smallest amounts. Caps and per-withdrawal limits start at the pilot values, not the target values.
   Raise them only after a clean week at each step, by the documented k-vote or install parameter.
4. Publish: INDEX SCID, factory SCID, template hashes, signer set, k/n, bonds, caps, viewing keys, and the custody
   disclosure in one signed post. No "trustless", "non-custodial", "bridge" or "fully backed" language.
5. Keep the previous release installable. A new release is a new INDEX; there is no update path on mainnet.

**Exit:** G8 and G9.

## 5. Critical path

```
recover wm/wz/wv
      │
      ├── vault v4 freeze ──► external review ──► remediation ──┐
      │                                                          │
      └── Monero multisig coordinator ──► stagenet 2-of-3 ───────┤
                                                                 ├──► pilot ──► mainnet rehearsal ──► launch
          Zcash testnet pilot ───────────────────────────────────┘
```

The two items that can move the date most are (1) recovering the `wm` work and (2) the external review queue. Start
both in week 1. Do not sequence the audit after the pilot.

## 6. First ten working days

1. Find and commit the three worktrees; push them to the fork. If they cannot be found, declare a rebuild and
   re-baseline the schedule.
2. Create `release/wrap-launch` from `origin/main`; merge the signer retry commit; keep other work separate.
3. Restore `vendor/derohe`, Go 1.26.8 and Node; run `tools/check.sh`; save the baseline log.
4. Re-check the matrix's cites against the merged tree; mark each row `implemented`, `partial`, `missing` or
   `stale`.
5. Stand up the WT work board with owners and test names; put the 19 critical rows first.
6. Reproduce WT-69 and WT-17 on the current app; fix them first because they block the only permissionless exit.
7. Reproduce WT-01 and WT-03 on the simulator; these decide whether any v4 withdrawal can be paid.
8. Write the Monero coordinator protocol: round names, message shapes, verification, cancel path, retry rules.
9. Send the external review package for the frozen contract candidate.
10. Open the operator decision sheet for the launch-blocking choices in §8.

## 7. Workstreams

| Workstream | Accountable | Deliverable |
|---|---|---|
| Contracts and vault | contracts engineer | v4 frozen, audited, DVM suite green, parameters table |
| Monero | Monero engineer | multisig coordinator, stagenet E2E, fee and consolidation rules |
| Zcash | Zcash engineer | testnet E2E, consolidation, encrypted seed, v6 deposits |
| Daemon and coordination | daemon engineer | crash-safe passes, health/metrics, `vaultctl`, active-set handling |
| App | app engineer | every app WT row closed; mainnet-shaped UI pass with a vault |
| Operations | operator + ops | named signers, hosts, nodes, ceremony, backups, on-call, pilot |
| Assurance | external reviewer | contract and backend review; remediation sign-off |
| Release | coordinator | reproducible builds, signed artifacts, deployment record |

## 8. Operator decisions that block the date

These are not engineering tasks and cannot be inferred:

- **YB-3:** the named signers, their hosts, k and n. Rule: n ≥ 2k.
- **YB-7:** `cap`, `capratio`, `bigbps`, `bigdelay`, `winmax`, `wdelay`, and the bond-to-cap table.
- **YB-18:** who pays the first DERO gas for a holder arriving with only wXMR or wZEC.
- **YB-20:** who signs `vaultsigner` releases, and from which canonical repository.
- **YB-23:** whether vault fees reach pass holders, and the monotone counter that would require.
- **YB-25:** what "a market the wrapped tokens can be traded on 1:1" means, if anything.
- **Treasury:** the mainnet treasury addresses and the bot's funding.
- **Assurance:** the external reviewer and the budget.
- **Legal and communications:** the custody disclosure, jurisdiction, and who speaks for the project.
- **Repository:** write access to the canonical repo, or a decision that the fork is the release repo.

## 9. Definition of production grade

A launch is not "the tests pass". All of these must hold:

- **Safety:** no open critical or high acceptance row; exactly-once payment proven under crash, restore and
  equivocation; reserve invariant checked continuously and independently.
- **Liveness:** a signer restart, node restart or single signer outage does not stop withdrawals; a halt lapses
  without a human; Reclaim is reachable from the app.
- **Recoverability:** restore from seed and from backup rehearsed; no double payment after a restore; migration
  from a leaked share rehearsed.
- **Observability:** health and metrics per signer; alerts that fire before a quorum is lost; a daily reserve
  report.
- **Change control:** frozen commits, reproducible builds, signed artifacts, a published hash for every contract
  and binary.
- **Operations:** named on-call, incident runbooks with time budgets, a communication channel, and a rollback plan
  that does not require mutating an immutable contract.
- **Disclosure:** custody, privacy (transparent ZEC is public), fees, caps and failure modes stated before a user
  deposits.
- **Capacity:** measured throughput per chain, a withdrawal queue bound, and a published maximum withdrawal size
  given the 200-input and fee constraints.

## 10. Top risks

| Risk | Impact | Mitigation |
|---|---|---|
| The `wm`/`wv` work is lost | Months of rework; the Monero launch slips | Recover in week 1; if lost, rebuild from §5.3 and the matrix, and re-baseline publicly |
| Monero multisig double-payment or nonce reuse | Direct loss, loss of confidence | WT-04 to WT-08, WT-10, WT-56 and WT-71 tests before any real funds; anchor-input rule; cancel path |
| Audit finds a critical defect late | Launch slips or ships unsafe | Start review at G2; internal adversarial rounds first (Y5-1); keep a schedule buffer |
| Fee accounting mismatch | Vault halts on first withdrawal or leaks the fee pool | WT-01 on chain, per-chain regtest with real fees, `reserve == sup − inflight` after each |
| Deposit attribution after a state loss | An honest deposit never mints | WT-11 on-chain anchor; restart drills with an empty state |
| Signer recruitment and key ceremony | No federation exists to launch | Operator starts now; dry-run ceremony on testnet; n ≥ 2k |
| Node or wallet compromise (WT-67, WT-68) | Forged votes or a drained fee wallet | Loopback + login enforcement; independent nodes per signer; no shared operators |
| Toolchain or provenance drift | Artifact cannot be reproduced or trusted | Pin Go 1.26.8, vendor, Node; full clone; signed release manifest |
| Regulatory or communications error | The product is misdescribed as trustless | Operator-approved disclosure; forbidden-words check in CI |

## 11. Launch-day outline

- **T-14:** freeze; final `check.sh`; publish candidate hashes; confirm signers, nodes and funding.
- **T-7:** mainnet rehearsal on testnet; incident drill; verify every runbook step against the actual flags.
- **T-1:** fund the deployer and treasury; confirm wallet logins; confirm the signer set is seated and bonded;
  confirm the on-call rota and the communication channel.
- **T-0:** install contracts; install and seal the vaults; `SetViewKey`; `verifydeploy` with no `-allow`; install
  the app on a ring-16 INDEX; open at the lowest caps.
- **T+1 to T+7:** daily reserve proof, daily signer health, one small withdrawal per asset, no cap increase.
- **T+7:** review the week's evidence. Raise a cap only if every day was clean and the reserve invariant held.
- **Any halt, unexplained shortfall, or double-payment signal:** stop cap increases, follow
  `INCIDENT_RESPONSE.md`, publish what is known. Do not improvise a contract call.

## 12. The next decision

The plan cannot start honestly until the operator answers two questions:

1. **Where are the `wrap-monero`, `wrap-zcash` and `wrap-vault` worktrees?** If they exist, this plan is an
   integration and assurance plan. If they do not, it is a rebuild plan, and the date changes.
2. **Who are the signers, and what are k and n?** Without named third parties there is no federation to launch,
   only software that could run one.
