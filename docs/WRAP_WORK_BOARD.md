# DMZ wrap work board

*2026-11-09. Generated from `docs/WRAP_ACCEPTANCE_MATRIX.md` (74 rows: 19 critical, 28 high, 20 medium, 7 low). The
matrix is the source of truth for the threat, the current protection, the required mitigation and the full required
tests. This board is the execution view: one row per acceptance item, one accountable owner, one status, the named
tests, and the directory where evidence must land.*

## How to read it

- **Owner** is the first owner named in the matrix. The matrix says the first owner is accountable; the other named
  owners are supporting owners, and their parts stay in the matrix row.
- **Tests** lists named Go/JS tests where the matrix names them. Where the matrix names no test, it lists the required
  verification type instead (DVM, Regtest, Simulator e2e, Unit, Fault injection, uitest, TLC, ...). A verification
  type is a plan, not an acceptance: name the test or drill in the evidence README before closing the row.
- **Evidence** is `research/wrapper/evidence/WT-XX/`. The directory does not exist yet. Create it when the first
  artifact lands; the required contents and the no-secrets rule are in `research/wrapper/evidence/README.md`.
- **Code state** is a separate audit against the merged tree: `docs/WRAP_CITATION_TRIAGE.md` marks all 74 rows
  `implemented` / `partial` / `missing` / `stale` (currently 0 / 47 / 27 / 0). It does not replace the status below.
- **Status** starts at `open` for every row. The matrix's Current protection column describes code, not accepted
  risk; a row is accepted only when its required mitigation and required tests exist.

## Status rules

| Status | Meaning |
|---|---|
| `open` | No accepted mitigation and no passing test. This includes rows whose current code already looks right. |
| `in-progress` | An owner has started; the named test or verification plan is being written. |
| `blocked` | Waits on a named item, decision or external dependency. Record the blocker in the row's evidence README. |
| `done` | Every required test passes, and the evidence directory holds the command, log, exit status and chain evidence. A code review is not evidence. |
| `accepted-risk` | The operator accepted the risk in writing. Only non-critical rows, and never used silently to pass a launch gate. |

## Counts

| Severity | Rows |
|---|---|
| critical | 19 |
| high | 28 |
| medium | 20 |
| low | 7 |
| **Total** | **74** |

| Test form | Rows |
|---|---|
| At least one named Go/JS test | 38 |
| Verification type only; no named test in the matrix | 36 |

## Load by accountable owner

| Owner | Critical | High | Medium | Low | Total |
|---|---|---|---|---|---|
| coordination | 8 | 12 | 4 | 3 | **27** |
| monero-backend | 7 | 3 | 4 | 1 | **15** |
| app-hub | 1 | 7 | 2 | 2 | **12** |
| vault-contract | 1 | 4 | 4 | 1 | **10** |
| zcash-backend | 1 | 1 | 4 | 0 | **6** |
| ops | 1 | 1 | 2 | 0 | **4** |

## Board

The table is sorted critical → high → medium → low. The first 19 rows are the launch-critical set; none of them can
be open at G3.

| ID | Severity | Threat | Owner | Status | Test(s) / verification | Evidence |
|---|---|---|---|---|---|---|
| WT-01 | critical | The foreign network fee leaves the reserve but not sup, and each backend handles it differently, with nothing recorded on chain | vault-contract | in-progress | TestReleasedChargesTheNetworkFeeToTheFeePool, TestNetworkFeeDoesNotHalt (also: DVM, Regtest, Unit) | `research/wrapper/evidence/WT-01/` |
| WT-02 | critical | A Paying mark lapses on a valid destination: v3 freezes the withdrawal even with a live quorum; v4 opens Reclaim at mark+20000 after the coins were paid if co-signers ne… | coordination | open | TestALapsedMarkWithAValidDestinationIsSettledWithoutAHuman, TestReclaimNeverOpensOnAPaidWithdrawalWhileKHonestAreLive, TestALapsedMarkKeepsTheIntentPaidRatherThanPayingAgain, TestNoPaymentStartsInsideTheMarkMargin (also: Fault injection, DVM, Unit) | `research/wrapper/evidence/WT-02/` |
| WT-03 | critical | The gas estimator runs contracts at height 0. On v4 every Paying is refused, so nothing is ever paid; after a halt lapses without Halt(0), every Attest and Mint is refus… | coordination | open | verification: Simulator e2e, Unit | `research/wrapper/evidence/WT-03/` |
| WT-04 | critical | A complete signed payment is held back, then broadcast after Veto or Reclaim returned the units: the withdrawal is paid twice | monero-backend | open | TestEveryAttemptForJSpendsTheAnchorInput (also: Regtest, Unit) | `research/wrapper/evidence/WT-04/` |
| WT-05 | critical | A proposer equivocates, or a second quorum signs, so two different payments are signed for one withdrawal | monero-backend | open | TestEquivocatingProposerCannotGetTwoTxsSignedForOneWithdrawal, TestEquivocatingProposerIsDetected (also: Unit) | `research/wrapper/evidence/WT-05/` |
| WT-06 | critical | After -verify-tries the daemon calls Release again while the first payment can still confirm | coordination | open | TestAPaymentThatConfirmsLateIsNotPaidAgain (also: Unit, Regtest) | `research/wrapper/evidence/WT-06/` |
| WT-07 | critical | A signer with lost or restored state, or two daemons on one identity, pays twice or drops a deposit | monero-backend | open | verification: Regtest, Fault injection, Unit | `research/wrapper/evidence/WT-07/` |
| WT-08 | critical | A Monero multisig wallet restored from a file copy signs again with nonces already used, and leaks its key share | monero-backend | open | verification: Regtest | `research/wrapper/evidence/WT-08/` |
| WT-09 | critical | Change, or a payment back to the vault, is minted as a deposit | monero-backend | open | TestDepositsIgnoreOutputsOfVaultSpends, TestValidDestRefusesEveryVaultAddress (also: Regtest, Unit) | `research/wrapper/evidence/WT-09/` |
| WT-10 | critical | A co-signer signs a transaction that pays elsewhere, adds outputs, sends change away, inflates the fee or sets a lock | monero-backend | open | TestSignerRefusesADifferingDigest, TestExpiryRequired (also: Unit, Regtest) | `research/wrapper/evidence/WT-10/` |
| WT-11 | critical | Each signer takes "address live since" from its own first sight, so an honest deposit gets fewer than k votes and is lost | coordination | open | TestLiveFromIsAFunctionOfChainData (also: Unit, Fault injection, Regtest, DVM) | `research/wrapper/evidence/WT-11/` |
| WT-12 | critical | The attested address is never checked against the signer's own derivation; Attest can re-attest; Mint needs no attested address. (critic) The one-shot rule and TestAttes… | coordination | open | TestAttestedAddressCannotBeReplaced, TestMintRefusedOnUnattestedIndex, TestDaemonHaltsOnAttestedAddressMismatch (also: DVM, Unit, Regtest, Runbook review) | `research/wrapper/evidence/WT-12/` |
| WT-13 | critical | Monero outputs the reserve cannot spend are minted: a duplicated one-time key, or a locked output | monero-backend | open | verification: Regtest, Unit | `research/wrapper/evidence/WT-13/` |
| WT-14 | critical | Zcash coinbase paid to a deposit address is minted and counted in the reserve, but a transparent-only vault cannot spend it | zcash-backend | open | TestRegtestCoinbaseToDepositAddressIsNeverMinted (also: Regtest) | `research/wrapper/evidence/WT-14/` |
| WT-15 | critical | The reserve check is blind: unminted deposits count as backing, unpaid burns count as in flight, stale Monero balances pass, and a quorum's Monero spend is invisible. (c… | coordination | open | verification: Unit, Fault injection, Regtest | `research/wrapper/evidence/WT-15/` |
| WT-16 | critical | The components' rules on signer-set shape contradict each other | ops | open | TestRedeemScriptFitsForEveryAllowedN, TestHealthRefusesWhenContractKOrNaDiffers (also: DVM, Unit, Doc lint) | `research/wrapper/evidence/WT-16/` |
| WT-17 | critical | FE-3: a successful Burn or Register is reported as "did not execute", and the address, the withdrawals and Reclaim never show, so users burn again | app-hub | in-progress | tests/js/vaultkeys.test.mjs; DVM; uitest | `research/wrapper/evidence/WT-17/` |
| WT-21 | critical | A false or reorged mint, completed by k−1 liars and one eclipsed, lazy or copying signer, goes unpunished. (critic) TestExecutedMintWithoutDepositTriggersHaltAndSlash … | coordination | open | TestExecutedMintWithoutDepositTriggersHaltAndSlash, TestUnbackedMintThenBurnIsNeverPaid, TestMintVoteKeysIdentifyEveryVoter (also: DVM, Simulator, Fault injection) | `research/wrapper/evidence/WT-21/` |
| WT-66 | critical | k signers' DERO wallets alone, with every foreign key untouched, vote Mint on a deposit that does not exist. They burn the units in the next block, and one of those keys… | coordination | open | TestUnbackedMintThenBurnIsNeverPaid (also: Simulator, Regtest) | `research/wrapper/evidence/WT-66/` |
| WT-18 | high | The DERO signer set, the foreign key shares and the coordination pins drift apart on Slash, Retire or AddSigner2 | coordination | open | TestMessagesFromARemovedSignerAreDropped, TestNoForeignActionWhenNotActive, TestDaemonRefusesWhenCustodyDescriptorDiffers (also: Regtest, Unit) | `research/wrapper/evidence/WT-18/` |
| WT-19 | high | Bonds do not deter a quorum: about 0.6 DERO stands against 100 XMR or 1,000 ZEC; v4 Retire returns a bond at once; slashed bonds go to the treasury | vault-contract | open | TestRetireWaitsForTheUnbondingDelay, TestARetiringSignerIsStillSlashable, TestQuorumCannotRecoverBondsBySeatingPuppets, TestSlashedBondNeverReachesTheTreasury, TestParametersBondValueCoversCap, TestVaultTreasuryIsNotASignerOrOwner (also: DVM) | `research/wrapper/evidence/WT-19/` |
| WT-20 | high | winmax does not bound a colluding or buggy quorum; 0 means unlimited; Tighten is one-way; the limits can lose the race to Seal | vault-contract | open | TestSealRefusesUnlimitedMintWindow, TestSealRefusesUnconfirmedLimits, TestAQuorumCannotMintMoreThanWinmaxPerWindow, TestLargeDepositEventuallyMints (also: DVM) | `research/wrapper/evidence/WT-20/` |
| WT-22 | high | Finality is a flat 10 confirmations in the daemon, and nothing reacts when a credited deposit or a Released payment vanishes. (critic) Monero deposits are credited at 10… | coordination | open | TestReleasedPaymentReorgedOut, TestConfirmationTierByAmount, TestMainnetMoneroFloorCoversTheRecordedReorgBound (also: Regtest, Fault injection, Unit) | `research/wrapper/evidence/WT-22/` |
| WT-23 | high | The daemon pays on a mark read at one confirmation, with no margin before the mark lapses | coordination | open | TestPayWaitsForAStableMark, TestNoPaymentStartsWithLessThanTheMarginLeft (also: Simulator) | `research/wrapper/evidence/WT-23/` |
| WT-24 | high | One foreign read error, a fatal RPC exit, or full-state reads that keep growing stop every withdrawal | coordination | open | TestDepositScanFailureDoesNotBlockWithdrawals, TestPassCostIsIndependentOfStorageSize (also: Unit, Fault injection) | `research/wrapper/evidence/WT-24/` |
| WT-25 | high | Spam and reverted votes drain signer DERO wallets and stretch a pass past a 2000-block bucket | coordination | open | TestVetoPassDoesNotRevoteAStandingVote, TestVetoWaitsForAStableBurn (also: Unit, Simulator, DVM) | `research/wrapper/evidence/WT-25/` |
| WT-26 | high | Honest daemons vote a compromise halt back down; one honest misreading plus k−1 standing Halt(1) votes halts for 20,000 blocks, and a vote cannot be withdrawn | coordination | open | TestNoResumeVoteIsEverCast, TestTransientShortfallDoesNotVoteHalt, TestAStandingHaltVoteCanBeWithdrawn (also: Unit, DVM, Simulator) | `research/wrapper/evidence/WT-26/` |
| WT-27 | high | One signer's Paying mark makes honest daemons start a new payment while halted in a shortfall | coordination | open | TestAForeignMarkDoesNotStartAPaymentWhileHalted (also: Unit) | `research/wrapper/evidence/WT-27/` |
| WT-28 | high | Coordination identity is self-asserted, removed signers stay pinned, and old attempts can replay | coordination | open | TestIdentityNotSignedByTheDeroKeyIsRefused, TestPinnedSetEqualsActiveSetOnChain, TestMessagesFromARemovedSignerAreDropped, TestAReplayedPreviousAttemptIsNotActedOn (also: Unit, Regtest, Fault injection) | `research/wrapper/evidence/WT-28/` |
| WT-29 | high | A wrongly built Zcash P2SH script, signature, fee or branch id locks or splits custody | zcash-backend | open | verification: Unit, Regtest, Property | `research/wrapper/evidence/WT-29/` |
| WT-30 | high | Monero ceremony and node trust: wrong threshold or version, a substituted participant, a remote or unauthenticated monerod or witness, a stale set_daemon | monero-backend | open | TestBackendRefusesRemoteMonerodWithoutWitness, TestBackendRefusesWalletRpcWithoutLogin (also: Regtest, Fault injection, Unit) | `research/wrapper/evidence/WT-30/` |
| WT-31 | high | The deposit id's spelling or vout meaning differs between builds or chains, so one output mints twice or a deposit never reaches quorum | coordination | open | TestDepositIDIsExplorerOrderLowercase, TestNoMintWhenTheReversedIDIsClosed, TestVaultDepositIdIsCanonical (also: Unit, Regtest) | `research/wrapper/evidence/WT-31/` |
| WT-32 | high | App, backends and daemon versions disagree on which destinations are valid, so a paid withdrawal is vetoed or an unpayable one is paid | coordination | open | TestValidDestPerChainAndNetwork, tests/js/dest.vectors.test.mjs, TestAnIntentWhoseDestinationTheChainRejectsEndsInVeto (also: Go and JS, Unit, Fault injection) | `research/wrapper/evidence/WT-32/` |
| WT-33 | high | One Monero wallet reused across vaults (v3 → v4) mints one deposit on both | monero-backend | open | TestBackendRefusesAWalletUsedByAnotherVault (also: Unit, Regtest) | `research/wrapper/evidence/WT-33/` |
| WT-34 | high | The app shows whatever deposit address one node returns, with no derivation, network or checksum check | app-hub | open | tests/js/derive.vectors.test.mjs (also: JS, uitest) | `research/wrapper/evidence/WT-34/` |
| WT-35 | high | Reclaim and older withdrawals fall off the screen, v3 lists none of this wallet's withdrawals, and the tracker copy contradicts v4 | app-hub | open | TestDaemonReclaimsAnUnmarkedExpiredBurn (also: Simulator e2e, uitest, Unit) | `research/wrapper/evidence/WT-35/` |
| WT-37 | high | The app claims more than the contract records: mint "Done" from the wallet balance, "Settled on Monero", "Waiting" for an unseen payment | app-hub | open | verification: uitest | `research/wrapper/evidence/WT-37/` |
| WT-38 | high | A retired vault still takes deposits, and the vault's code is not hash-pinned | app-hub | open | verification: uitest | `research/wrapper/evidence/WT-38/` |
| WT-39 | high | Wrapped units are labelled "XMR"/"ZEC" on DERO, and near-miss impersonator tickers pass | app-hub | open | verification: uitest, JS | `research/wrapper/evidence/WT-39/` |
| WT-40 | high | Reserve-proof and privacy copy is false: a green tick for any stored key; a Monero view key cannot see spends; transparent ZEC makes deposits public | app-hub | open | verification: uitest, Regtest, Unit | `research/wrapper/evidence/WT-40/` |
| WT-41 | high | Deployment order: no foreign keys committed before Seal; Register opens before a viewing key; a bundle can name an unsealed or unaudited vault | vault-contract | open | TestSealRefusesWithoutEveryForeignKey, TestForeignKeyWritableOnlyByItsSigner, TestRegisterWaitsForCustody, TestHealthFailsWhenWalletIsNotTheCommittedCustody, TestMainnetBundleRefusesUnsealedOrUnauditedVault (also: DVM, Unit, Regtest) | `research/wrapper/evidence/WT-41/` |
| WT-47 | high | Incident response has no tooling: every k-vote is hand-built on mainnet, the compromise procedure is incomplete, and the drill's reserve sweep has no code path | ops | open | verification: Simulator, Regtest | `research/wrapper/evidence/WT-47/` |
| WT-67 | high | The DERO signer wallet-rpc is never required to ask for a login or to be on loopback. Any local process or LAN host can cast this signer's Mint, Paying, Halt(0) and Rele… | coordination | open | TestMainnetSignerRefusesAWalletThatAsksForNoLogin, TestMainnetSignerRefusesARemoteWallet, TestHealthFailsWhenTheWalletStopsAskingForALogin (also: Unit) | `research/wrapper/evidence/WT-67/` |
| WT-68 | high | Each signer's chain views come from nodes nobody cross-checks. Suppose the signers use one hosted DERO node or one public monerod. If it serves a forged w_7_to, every co… | coordination | open | TestMainnetRefusesASingleZcashNode, TestMainnetRefusesMoneroWithoutAWitness, TestPaymentReadsTheWithdrawalFromTwoNodesAtStableHeight (also: Fault injection, Unit) | `research/wrapper/evidence/WT-68/` |
| WT-69 | high | The app estimates height-gated calls, and the estimator runs contracts at BLOCK_HEIGHT() = 0. Reclaim therefore always "would fail", so the only permissionless exit cann… | app-hub | in-progress | tests/js/wrap.heightgate.test.mjs; Simulator e2e | `research/wrapper/evidence/WT-69/` |
| WT-70 | high | Custody migration has no implementation. Suppose one Monero key share leaks (fewer than k). The incident runbook says to migrate the reserve now, but no sweep tool exist… | vault-contract | open | verification: Regtest, DVM | `research/wrapper/evidence/WT-70/` |
| WT-71 | high | The Monero backend adopts a payment the contract already recorded for another withdrawal. A user burns 1 XMR to D twice (j1, j2), and j1 is paid by T1 while signer A is … | monero-backend | open | TestAPaymentReleasedForAnotherWithdrawalIsNotAdopted (also: Unit, Regtest) | `research/wrapper/evidence/WT-71/` |
| WT-36 | medium | The confirmation sheet before Register, Reclaim and Burn uses a stale vault read, quotes a delay instead of heights, and does not label the destination | app-hub | open | verification: uitest, Unit | `research/wrapper/evidence/WT-36/` |
| WT-42 | medium | The minimum applies per output, so split payments under mindep are lost, and four documents disagree about what happens | coordination | open | verification: Regtest, uitest, JS | `research/wrapper/evidence/WT-42/` |
| WT-43 | medium | Over-cap deposits wait for ever and can starve; the app overstates the room to mint | coordination | open | verification: Unit, uitest, Regtest | `research/wrapper/evidence/WT-43/` |
| WT-44 | medium | Deposits in limbo (under minimum, before the anchor, unregistered or unattested index) need a human | zcash-backend | open | verification: Regtest | `research/wrapper/evidence/WT-44/` |
| WT-45 | medium | Decimals, chain and confirmation policy come from unverified config, so the app shows the wrong minimum and count | app-hub | open | verification: DVM, Unit, JS, uitest | `research/wrapper/evidence/WT-45/` |
| WT-46 | medium | v4 changed capratio to foreign units per whole DERO, but the docs, verifydeploy and the app still use per-atomic-unit arithmetic | ops | open | verification: Unit, JS | `research/wrapper/evidence/WT-46/` |
| WT-48 | medium | A failing signer goes quiet, and nothing alerts before a quorum is lost | coordination | open | TestUnhealthyIsVisibleOnTheEndpoint (also: Unit, Simulator) | `research/wrapper/evidence/WT-48/` |
| WT-49 | medium | The network is self-declared, so the file backend can run against a public testnet | coordination | open | verification: Unit | `research/wrapper/evidence/WT-49/` |
| WT-50 | medium | Released's ftx is free text; one foreign transaction settles one withdrawal; the payment-to-j binding lives only in the backends | vault-contract | open | TestReleasedTxidIsCanonical, TestVerifyReleaseRequiresTheJBinding (also: DVM, Unit, Regtest) | `research/wrapper/evidence/WT-50/` |
| WT-51 | medium | Splitting a burn escapes bigdelay | vault-contract | open | TestSplitBurnsStillEarnTheLargeDelay (also: DVM) | `research/wrapper/evidence/WT-51/` |
| WT-52 | medium | Governance votes never expire | vault-contract | open | verification: DVM | `research/wrapper/evidence/WT-52/` |
| WT-53 | medium | Monero pays one withdrawal at a time; a queue longer than the mark window strands withdrawals | monero-backend | open | TestReserveCountsLockedChange (also: Regtest, Unit) | `research/wrapper/evidence/WT-53/` |
| WT-54 | medium | A hard fork the build does not know (Monero FCMP++, a Zcash upgrade) stalls the signer or puts it on a dead chain | monero-backend | open | verification: Unit | `research/wrapper/evidence/WT-54/` |
| WT-55 | medium | Deposits from current Zcash wallets (v6 with shielded parts) cannot be verified | zcash-backend | open | verification: Unit, Regtest | `research/wrapper/evidence/WT-55/` |
| WT-56 | medium | A co-signer crashes between signing and recording | monero-backend | open | verification: Unit | `research/wrapper/evidence/WT-56/` |
| WT-57 | medium | After a restore, the Monero subaddress lookahead misses deposits | monero-backend | open | verification: Regtest | `research/wrapper/evidence/WT-57/` |
| WT-58 | medium | Operator documents contradict the contract: halt scope, parameter mutability, lock file, gas, cites, finality, custody | ops | open | verification: Parse test, TLC | `research/wrapper/evidence/WT-58/` |
| WT-59 | medium | Storage-gas headroom for v4 plus these fixes is unmeasured | vault-contract | open | verification: DVM | `research/wrapper/evidence/WT-59/` |
| WT-72 | medium | A fragmented reserve and fixed fee bounds make withdrawals unpayable, or charge one user for other people's deposits. Someone pays 400 outputs just above the per-input c… | zcash-backend | open | verification: Regtest, Unit, DVM, uitest | `research/wrapper/evidence/WT-72/` |
| WT-73 | medium | The Zcash signing seed is plaintext hot material, and every vault key sits under a published xpub. A backup, a disk image or any process running as the signer's user can… | zcash-backend | open | TestMainnetRefusesAPlaintextSeed, TestNoExportedFunctionReturnsAChildPrivateKey (also: Unit, Runbook review) | `research/wrapper/evidence/WT-73/` |
| WT-60 | low | Wrapped units attached to any entrypoint but Burn are destroyed | vault-contract | open | verification: DVM | `research/wrapper/evidence/WT-60/` |
| WT-61 | low | The daemon has no Height > 0 guard on deposits | coordination | open | verification: Unit | `research/wrapper/evidence/WT-61/` |
| WT-62 | low | An index is silently truncated | monero-backend | open | verification: Unit | `research/wrapper/evidence/WT-62/` |
| WT-63 | low | A return to an unregistered w_by fails, and Veto is re-voted for ever | coordination | open | verification: Unit | `research/wrapper/evidence/WT-63/` |
| WT-64 | low | "Claims outstanding" includes in-flight and fee-pool units without saying so | app-hub | open | verification: Unit | `research/wrapper/evidence/WT-64/` |
| WT-65 | low | The hub ships without its guards and tests | app-hub | open | verification: JS, uitest | `research/wrapper/evidence/WT-65/` |
| WT-74 | low | seatedSigners reads at most 64 seats, and n counts every seat ever filled. Once enough signers have been replaced that n exceeds 64, a newly seated signer reads as "neve… | coordination | open | verification: Unit | `research/wrapper/evidence/WT-74/` |
