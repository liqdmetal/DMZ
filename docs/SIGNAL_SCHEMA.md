# DERO signal schema, v1

**Status:** implementation schema for the signal-first DEX (`docs/SIGNAL_DEX.md`, work order 1).
**Code:** `tools/signal` — Go types, strict parser, canonical serializer and validator tests — and
`app/signal.js`, the browser mirror loaded by `app/index.html`. Both read
`tools/signal/testdata/signal_v1_vectors.json`, so a record accepted by one is accepted by the other.
**Scope:** public discovery data only. A valid signal is not proof of NFT ownership, approval, or
settlement. The taker still verifies the source chain before signing.

## Record

A signal is one JSON object with exactly these fields, in this order. `v` is the schema version; a
reader that does not know the version must reject the record, never guess.

| Field | JSON type | Meaning |
|---|---|---|
| `v` | integer | schema version; exactly `1` |
| `chain` | string | NFT's native chain: `eth` or `sol` |
| `asset` | string | collection/contract address on that chain |
| `token` | string | Ethereum token id, or Solana mint address |
| `side` | string | `sell` or `buy` |
| `price` | string | positive integer price in the payment asset's base units |
| `pay` | string | `native`, or a payment-asset address on the same chain |
| `expiry` | string | `time:<unix-seconds>` or `height:<source-chain-height>` |
| `maker` | string | address that can settle on the source chain |
| `nonce` | string | 32-byte lowercase hex, unique per maker and chain |
| `adapter` | string | settlement adapter and version, `<name>/<major>.<minor>.<patch>` |
| `ref` | string | 32-byte lowercase hex of the full terms document, including any SporeM3 payload |

Example (Ethereum sell):

```json
{"v":1,"chain":"eth","asset":"0x1111111111111111111111111111111111111111","token":"0","side":"sell","price":"1000000000000000000","pay":"native","expiry":"time:2000000000","maker":"0x2222222222222222222222222222222222222222","nonce":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","adapter":"seaport/1.6.0","ref":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}
```

## Canonical serialization

Canonical bytes are what two readers hash or compare when they must agree on a signal:

- JSON object, fields in the order above, no whitespace.
- Decimal integer strings, no sign, no exponent, no leading zero (except `"0"`).
- Addresses and hashes lowercase; Ethereum addresses include `0x`.
- No unknown fields, duplicate keys, trailing data, or alternate spelling of a field value.
- `Canonical()` in `tools/signal` and `Veil.signal.canonical()` in `app/signal.js` emit these bytes.
  Reordered or pretty-printed input produces the same canonical bytes after parsing.

`ref` is not the hash of this JSON object. It commits to the full terms document (including private
SporeM3 terms), so `ref` is supplied by the maker and checked only for shape here. A verifier that
has the terms document must recompute it separately.

## Per-chain values

### Ethereum (`chain: "eth"`)

- `asset`: `0x` followed by 40 lowercase hex digits; the zero address is invalid.
- `token`: canonical unsigned decimal, `0` to `2^256-1`; token id `0` is legal. `0x` token ids,
  leading zeros and signs are not canonical.
- `pay`: `native`, or an Ethereum address in the same lowercase form.
- `maker`: an Ethereum address in the same form; the zero address is invalid.

### Solana (`chain: "sol"`)

- `asset`, `token`, `maker`: base58 that decodes to exactly 32 bytes; the all-zero key is invalid.
  `token` is the mint address, not a decimal edition number.
- `pay`: `native`, or a base58 32-byte mint address.
- Solana addresses are case-sensitive and are not rewritten to Ethereum casing.

## Expiry

`expiry` carries its own discriminator so a timestamp can never be mistaken for a block height:

- `time:<seconds>` is Unix time. The signal is dead when `now >= expiry`; the value must be a
  positive Unix timestamp no greater than `math.MaxInt64`.
- `height:<blocks>` is a source-chain height. The signal is dead when `source height >= expiry`.
  A caller that does not know the source height must fetch it before treating the signal as live;
  passing height `0` to `Live` deliberately means "unknown, do not reject yet".

`tools/signal.Live(now, height)` and `Veil.signal.live(record, { now, height })` validate the record
and apply those two rules. Validation alone checks syntax and does not reject an expired record, so a
board can display history.

## Browser mirror

`app/signal.js` exposes the same operations under `Veil.signal`: `parse`, `validate`, `canonical`,
`parseExpiry` and `live`. Errors carry a `field` property naming the rejected field, which lets a
screen name the bad value without parsing an error string. It has no chain access and no adapter
allowlist; a screen must still verify ownership and settlement on the NFT's native chain.

The Go tests and `tests/js/signal.vectors.test.mjs` both consume the shared fixture. Add a case once,
then make both readers pass it; do not regenerate one side's expected result from its own output.

## Nonce and replay

`nonce` is 32 bytes of lowercase hex. It must be unique for the tuple `(chain, maker)`, and a maker
must never reuse one after cancellation, expiry, or a failed attempt. Uniqueness is enforced by the
publisher/board, not by this stateless validator. `nonce` and `ref` may not be all zero: that catches
an uninitialized field, which is otherwise a valid-looking hash.

## Adapter

`adapter` names the settlement implementation and its exact version, e.g. `seaport/1.6.0` or
`magiceden/2.0.0`. The name is lowercase letters, digits and hyphens, starting with a letter, at most
32 characters. The version is three non-negative decimal components. This validator checks the
shape; the app still needs an adapter allowlist before it can build a settlement transaction.

## Validation boundary

Accepted here:

- the record is structurally valid and canonically spelled;
- the address, token, price, hash and expiry shapes match the declared chain;
- the record has not reached its declared expiry (`Live`).

Not accepted as a result of that validation:

- ownership, approval, royalty, marketplace-order validity, or non-movement of the NFT;
- the maker controls `maker` or `asset`;
- `ref` matches the terms document;
- the adapter is deployed, safe, or supported;
- the signal can settle on the source chain.

Those are source-chain checks in the verifier, and a settlement is always the source chain's
transaction, not a DERO record.
