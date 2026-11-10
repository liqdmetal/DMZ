# WT-01 baseline evidence

| Field | Value |
|---|---|
| Row | WT-01 (critical): foreign network fee leaves the reserve but not `sup` |
| Accountable owner | vault-contract |
| Supporting owners | monero-backend, zcash-backend, app-hub |
| Status | `in-progress` — evidence started; mitigation and required tests are still missing |
| Date | 2026-11-10 |
| Code under test | `a417a8e01b0c0741df35d396ad2e2a44e51b3b60` (`origin/main`) |
| HEAD | `2401a5364c9d8c8b650650b12759f4270174d6e3` (docs-only work-board commit) |
| DERO harness | `e9df1205b6603c62f0651d0e18e5e77a2584b15e` (`deploy/vendor.lock`) |
| Toolchain | `go1.26.8 windows/amd64` |

## Verdict

**Not baselined — both required tests are absent at the code commit.**

`go test -list` against the real DVM finds `TestFeeUnitsStayInSupAndNeedTheTreasuryKeyForEver` and neither
`TestReleasedChargesTheNetworkFeeToTheFeePool` nor `TestNetworkFeeDoesNotHalt`. Running the two required names with
`-run` exits 0 because `go test` treats “no tests to run” as success; that exit status is not evidence of a pass.
The full output is in `test.log`.

The closest existing test, `TestFeeUnitsStayInSupAndNeedTheTreasuryKeyForEver`, **passes** and asserts the behaviour
WT-01 is meant to replace: after `ClaimFees`, `sup` is unchanged at 3510 and the 10 fee units remain in `sup`, below
`minburn`, so the treasury cannot redeem them by burning. The current suite therefore protects the old accounting; it
does not test the required v4 rule.

## Current code

- `contracts/vault.bas:406-424` — `Released(j, ftx)` has no `nfee` argument. It does `sup -= w_v` and
  `feepool += w_f`, so the network fee is not removed from `sup` and cannot be recorded on chain.
- `tools/vaultsigner/main.go:716-741` — the reserve invariant is `reserve >= sup - inflight` with no network-fee term.
- `tests/zz_adversary_priv_test.go:120-158` — the existing fee test confirms the old rule and is not a WT-01 test.

The required mitigation is not present: there is no v4 `Released(j, ftx, nfee)`, no `nfee <= w_f` refusal, and no
`sup -= w_v + nfee` / `feepool += w_f - nfee` accounting. `TestNetworkFeeDoesNotHalt` is also absent.

## Tests run

1. List the three relevant names against the copied `veil/tests` package. Only the existing test is present.
2. Run `TestFeeUnitsStayInSupAndNeedTheTreasuryKeyForEver`: **PASS**, exit 0.
3. Run the two WT-01 names: **no tests to run**, exit 0. This is a missing-test result, not a pass.

`tests/run.sh` copies `reference/`, `tests/` and `contracts/` into `vendor/derohe/veil` and runs there. The DERO
checkout is pinned and git-ignored; the one path `deploy/vendor.lock` records as deleted
(`vendor/github.com/tjfoc/gmsm/API????.md`, which Windows cannot create and no build uses) was excluded from the
checkout. No Monero, Zcash or network-fee regtest was run for this baseline.

## Reproduce

```sh
GOTOOLCHAIN=go1.26.8 tests/run.sh -list \
  'Test(FeeUnitsStayInSupAndNeedTheTreasuryKeyForEver|ReleasedChargesTheNetworkFeeToTheFeePool|NetworkFeeDoesNotHalt)$' -v

GOTOOLCHAIN=go1.26.8 tests/run.sh -run '^TestFeeUnitsStayInSupAndNeedTheTreasuryKeyForEver$' -v

GOTOOLCHAIN=go1.26.8 tests/run.sh -run \
  '^(TestReleasedChargesTheNetworkFeeToTheFeePool|TestNetworkFeeDoesNotHalt)$' -v
```

## Why the row is not done

The evidence records the missing test, not an accepted mitigation. To close WT-01, the v4 `Released` rule must be
implemented and both named tests must exist and pass, followed by the matrix’s 50-withdrawal regtest per chain with
real fees and the `reserve == sup - inflight` / no-Halt checks. Those results belong in the same directory.
