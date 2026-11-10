# WT-69 reproduction and app-side fix

| Field | Value |
|---|---|
| Row | WT-69 (high): app estimates height-gated calls at `BLOCK_HEIGHT() = 0` |
| Accountable owner | app-hub |
| Status | `in-progress` — app-side fixed and passing; simulator e2e still required |
| Date | 2026-11-10 |
| Code under test | `a55ce01` (code tree `a417a8e01b0c0741df35d396ad2e2a44e51b3b60`, `origin/main`) |
| Tests | `tests/js/wrap.heightgate.test.mjs`, registered in `tests/js/index.js` |
| Toolchain | Node `22.10.0`, Windows |

## Verdict

**Reproduced on the old code, then fixed.** The failing run is retained as `test.log`; the passing runs are
`test-pass.log` (focused) and `suite-pass.log` (full JS suite).

Old code, both tests failed:

```text
not ok 1 - WT-69: Reclaim is sent at a fixed fee, not refused by the height-0 estimator
  error: Reclaim must not be estimated: DERO.GetGasEstimate runs the contract at BLOCK_HEIGHT() 0
  1 !== 0
not ok 2 - WT-69: Register is sent at a fixed fee, not refused by the height-0 estimator
  error: Register must not be estimated: DERO.GetGasEstimate runs the contract at BLOCK_HEIGHT() 0
  1 !== 0
# tests 2
# pass 0
# fail 2
```

After the fix:

```text
1..2
# tests 2
# pass 2
# fail 0
exit=0
```

The full JS suite is `148` tests, `148` pass, `0` fail (`suite-pass.log`).

## Mitigation implemented

- `app/app.js` defines `HEIGHT_GATED_FEE = 5000` (0.05 DERO, well under the app's `MAX_FEE` of 100000) and a
  `HEIGHT_GATED` set containing `Reclaim`, `Register`, `Paying`, `Veto` and `Released`.
- In `app.tx`, when the caller passes no `opts.fee` and the entrypoint is in that set, the fixed fee is used and
  `DERO.GetGasEstimate` is never called. Non-gated calls keep the existing estimate and cap behavior.
- `doReclaim` and `doRegister` still pass no fee, so they now take the fixed-fee path without changing their call
  shape.

The tests assert the user-visible contract: zero estimates, exactly one wallet transfer with a positive fee within
the 1 DERO cap, and the wallet's txid returned. The old failure was not an estimate error that should be caught; the
call must not be estimated at all.

## Tests run

```sh
.tools/node-v22.10.0-win-x64/node.exe --test tests/js/wrap.heightgate.test.mjs
.tools/node-v22.10.0-win-x64/node.exe --test tests/js
```

## Not done

This is still a unit proof, not WT-69's required simulator e2e. Closing the row needs the real daemon estimator and
simulator: Reclaim executes past `rel + 20000` and moves `st` to 2; Register executes after a lapsed halt; on v4 a
never-attested index registers after `uh + 20000`; and a v4 daemon marks and pays a withdrawal. Those results belong
in this directory.
