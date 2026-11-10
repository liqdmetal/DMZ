# WT-69 reproduction

| Field | Value |
|---|---|
| Row | WT-69 (high): app estimates height-gated calls at `BLOCK_HEIGHT() = 0` |
| Accountable owner | app-hub |
| Status | `in-progress` — failing tests added; no fix |
| Date | 2026-11-10 |
| Code under test | `a417a8e01b0c0741df35d396ad2e2a44e51b3b60` (`origin/main`) |
| Tests | `tests/js/wrap.heightgate.test.mjs`, registered in `tests/js/index.js` |
| Toolchain | Node `22.10.0`, Windows |

## Verdict

**Reproduced.** Both new tests fail on the current code:

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

The full JS suite is 148 tests, 144 pass, 4 fail; the other two failures are WT-17.

For both entrypoints the app makes exactly one `DERO.GetGasEstimate` call, receives the height-0 refusal, never calls
`wallet.transfer`, returns `null`, and toasts `Reclaim would fail: …` / `Register would fail: …`. The tests assert the
opposite: zero estimates, one wallet transfer with a positive fee within the app’s 1 DERO cap, and the wallet’s txid
returned.

## Why

- `app/app.js:177-187` estimates every call that has no `opts.fee`.
- `app/wrap.js:339-352` (`doRegister`) and `app/wrap.js:355-365` (`doReclaim`) pass no fee.
- The real estimator runs the contract at `BLOCK_HEIGHT() = 0`, where Reclaim is before `rel + 20000` and Register
  after a lapsed halt is still halted, so the estimate always refuses.

The test node refuses only `DERO.GetGasEstimate`; the wallet would accept the same call. That isolates the app’s
estimate decision from the wallet and from a real daemon.

## Not done

This is a unit reproduction, not WT-69’s required simulator e2e. Closing the row still needs the real daemon estimator
and simulator: Reclaim executes past `rel + 20000` and moves `st` to 2; Register executes after a lapsed halt; on v4 a
never-attested index registers after `uh + 20000`; and a v4 daemon marks and pays a withdrawal. Those results belong
in this directory.
