# Wrap acceptance evidence

One directory per row of `docs/WRAP_ACCEPTANCE_MATRIX.md`, named by the row id:

```
research/wrapper/evidence/WT-01/
research/wrapper/evidence/WT-02/
...
```

The board that points here is `docs/WRAP_WORK_BOARD.md`. A row is `done` only when its directory holds the evidence
below. A code review, a passing unrelated suite, or an assertion in a document is not evidence.

## What each directory must contain

| File | Required when | Contents |
|---|---|---|
| `README.md` | always | Row id, accountable owner, status, the exact mitigation implemented, the commit under test, date, environment (OS, Go/Node/tool versions), and the commands run. Keep it short; link the rest. |
| `test.log` | a test or suite proves the row | The exact command, its full output, the exit status and the timestamp. Use a focused `go test -run ...` or `node --test ...` command, not only a whole-repository run. |
| `txids.txt` | the row needs chain evidence | Network, txid, block height, confirmations and the independent source that showed it. One line per transaction; no explorer-only claims. |
| `review.md` | the row is closed by a runbook review, drill, ruling or external review | Who reviewed, when, what they checked, what they found, and the sign-off. For a drill, record the timings the row requires. |

A row may need more than one of these. Name additional files by content (`fault-injection.log`, `rehearsal.md`), not
by date.

## Rules

- **No secrets.** Never commit a private key, seed, wallet password, RPC login, cookie, session token or API key. The
  Monero and Zcash material stays outside the repository; evidence names the fact that it was checked, never the
  material.
- **No machine-specific paths.** Use repository-relative paths and the command as run, not an absolute path from one
  host. If a log contains one, redact it before committing.
- **No personal data.** Redact addresses only when the row does not need them; never include a counterparty's
  identity, IP or hostname.
- **Reproducible.** A second person must be able to read `README.md`, run the command and get the same result against
  the same commit and fixtures. Record fixture hashes where the test uses one.
- **Failures stay.** If a test failed and was fixed, keep the failing log and add the passing one. Do not overwrite a
  failure with a success; the matrix requires the test to fail on the old code where possible.
- **Status changes go with evidence.** Update `docs/WRAP_WORK_BOARD.md` in the same commit that adds the evidence. A
  status change without a directory is a claim, not a record.
- **One row, one owner.** The first owner in the matrix row is accountable. Supporting owners add their evidence to
  the same directory and are named in its `README.md`.

## Creating a directory

```sh
mkdir -p research/wrapper/evidence/WT-01
$EDITOR research/wrapper/evidence/WT-01/README.md
```

Do not create empty directories in advance. The board's path is the target; the directory appears when the first
artifact lands.
