# WT-17 reproduction

| Field | Value |
|---|---|
| Row | WT-17 (critical): a successful Burn or Register is reported as “did not execute” |
| Accountable owner | app-hub |
| Supporting owners | vault-contract |
| Status | `in-progress` — failing tests added; no fix |
| Date | 2026-11-10 |
| Code under test | `a417a8e01b0c0741df35d396ad2e2a44e51b3b60` (`origin/main`) |
| Tests | `tests/js/vaultkeys.test.mjs` (two new v4 cases) |
| Toolchain | Node `22.10.0`, Windows |

## Verdict

**Reproduced for v4.** The two new tests fail on the current code:

```text
# vault version 4 is not the version this app reads
not ok 8 - WT-17: a v4 vault loads before its Register or Burn can be called
  error: the app must read a v4 vault; a v3-only read leaves v4 Burn and Register unavailable
not ok 9 - WT-17: v4's ADDRESS_STRING keys are visible to the Register and Burn probes
  error: uc_ is keyed by ADDRESS_STRING in v4, not by the 33 raw bytes
  0 !== 1
# tests 9
# pass 7
# fail 2
```

The seven existing v3 tests still pass. The failures are specifically the v4 path:

- `D.loadVault` refuses `v: 4` and returns `null`, so the app cannot call or report v4 Burn or Register.
- With the version gate bypassed, `D.vaultIndexes` asks for `uc_` plus the 33 raw bytes. A v4 contract stores
  `uc_<ADDRESS_STRING>`, so the read returns `0`; `D.vaultWithdrawals` compares `w_<j>_by` as raw hex while v4 stores
  the `dero1…` string, so the withdrawal is invisible.

Those are the reads that feed the Register and Burn probes. When they return no record after a successful call,
`app.tx` sets `executed = false` and shows the false “did not execute” toast. This test records the root cause; it
does not yet drive the toast through a rendered Wrap screen.

## What the tests set up

`v4State()` is the storage shape the matrix records for `wv:vault.bas:285-292, 414`:

- `kind = "vault"`, `v = 4`;
- `uc_<dero1> = 1`, `ul_<dero1>_0 = 0`;
- `w_0_by = dero1`, with the usual `w_0_v/f/h/rel/st` fields.

The node answers `GetSC` as `rpc_dero_getsc.go` does: `keysstring` for the string-keyed v4 read, `keysbytes` for the
app’s current raw read. No contract or app code was changed.

## Not done

WT-17 also requires `tests/vault_appkeys_test.go` against the real DVM and a uitest where a burn shows “confirmed”,
a `#j` row, and the index and address after Attest. Neither exists yet. The v4 app/contract must ship together; this
reproduction only makes the app-side gap executable.
