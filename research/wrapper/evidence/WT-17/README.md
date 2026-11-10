# WT-17 reproduction and app-side fix

| Field | Value |
|---|---|
| Row | WT-17 (critical): a successful Burn or Register is reported as “did not execute” |
| Accountable owner | app-hub |
| Supporting owners | vault-contract |
| Status | `in-progress` — app-side fixed and passing; DVM and uitest still required |
| Date | 2026-11-10 |
| Code under test | `a55ce01` (code tree `a417a8e01b0c0741df35d396ad2e2a44e51b3b60`, `origin/main`) |
| Tests | `tests/js/vaultkeys.test.mjs` |
| Toolchain | Node `22.10.0`, Windows |

## Verdict

**Reproduced for v4 on the old code, then fixed.** The failing run is retained as `test.log`; the passing runs are
`test-pass.log` (focused) and `suite-pass.log` (full JS suite).

Old code, both new v4 tests failed:

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

After the fix:

```text
1..9
# tests 9
# pass 9
# fail 0
exit=0
```

The full JS suite is `148` tests, `148` pass, `0` fail (`suite-pass.log`).

## Mitigation implemented

`app/vault.js` now selects the key encoding from the loaded vault version:

- `loadVault` accepts `v = 3` and `v = 4`; any other version still fails closed.
- v3 keeps the raw 33-byte encoding: `uc_` + raw, `ul_` + raw + `_c`, and `w_<j>_by` compared as raw hex.
- v4 uses `ADDRESS_STRING`: `uc_` + `dero1…`, `ul_` + `dero1…` + `_c`, and `w_<j>_by` decoded from the node's hex
  and compared as the `dero1…` text.
- `vaultIndexes`, `myVault` and `vaultWithdrawals` all use the same version-aware helpers, so the Register probe reads
  its own `uc_` record and the Burn probe reads its own `w_<j>_by` record in the contract's encoding.

`app/wrap.js`'s Burn probe now checks the new withdrawal itself rather than only a change in `wn`: it calls
`D.burnOpened`, which looks at the withdrawals opened since the call and requires `w_to == dest`,
`w_v == units − wfee`, and the wallet's own `w_<j>_by` in the version's encoding. A stale or wrong-key read can no
longer turn a successful burn into “did not execute”, which is the path that led users to burn again.

## Tests run

```sh
.tools/node-v22.10.0-win-x64/node.exe --test tests/js/vaultkeys.test.mjs
.tools/node-v22.10.0-win-x64/node.exe --test tests/js
```

`v4State()` is the storage shape the matrix records for `wv:vault.bas:285-292, 414`:

- `kind = "vault"`, `v = 4`;
- `uc_<dero1> = 1`, `ul_<dero1>_0 = 0`;
- `w_0_by = dero1`, with the usual `w_0_v/f/h/rel/st` fields.

The test now loads that vault through `loadVault` instead of bypassing the version gate, so the passing result covers
the real load path the Wrap screen uses. It also calls `D.burnOpened` directly for both v3 and v4 with the signed
destination and net units, and asserts a different amount is not accepted as this burn.

## Not done

WT-17 also requires `tests/vault_appkeys_test.go` against the real DVM and a uitest where a burn shows “confirmed”,
a `#j` row, and the index and address after Attest. Neither exists yet. This evidence closes only the app-side
reproduction; the row stays `in-progress` until those required checks pass.
