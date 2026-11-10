# Signal-first DEX direction

**Status:** operator direction, 2026-11-10. This is a design note, not a deployment spec.
**Decision:** DERO is the public signalling layer for now. SporeM3 is the fallback transport. NFT value and
settlement stay on the NFT's native chain, with Ethereum and Solana first. The wXMR/wZEC federation is parked as a
launch dependency.

## What DERO is, and is not

DERO can publish and timestamp small public records, and its contracts can hold a signal board without a server.
It cannot verify an Ethereum or Solana NFT owner, check an approval, or move a foreign NFT. A DERO signal is
therefore **advisory discovery data**, not proof of ownership and not settlement.

SporeM3 is a messenger. It can carry an encrypted negotiation or a private listing, but it is not a bridge, an
escrow, or a source of truth about a foreign chain. The repository's existing Spore line (`liqdmetal/spore`) is
already described this way; "SporeM3" here means the same role, not a new settlement layer.

## The model

```
signal (DERO, public)  ->  verify (ETH/SOL)  ->  settle (ETH/SOL)  ->  optional receipt (DERO)
        |                                                                    ^
        +---------------- SporeM3 private/fallback transport ----------------+
```

1. **Signal.** A maker publishes an intent: chain, collection/contract, token id, side, price, payment asset,
   expiry, maker, nonce, and the settlement adapter it expects. The public board stores the minimum needed to
   index it; sensitive terms can travel over SporeM3.
2. **Verify.** The taker's client reads the source chain directly and checks ownership, approval, royalties,
   expiry, and that the token has not moved since the signal was read. A stale or forged signal fails here.
3. **Settle.** The trade executes on Ethereum or Solana through that chain's own wallet and marketplace/program.
   DERO never holds the NFT or the purchase price.
4. **Receipt.** Optionally, DERO records a settlement receipt hash so the signal can be closed and the trade can
   be cited later. The receipt is not a claim on the asset.

## Signal records

Minimum public fields:

| Field | Meaning |
|---|---|
| `v` | signal schema version |
| `chain` | `eth`, `sol`, … |
| `asset` | collection/contract address on that chain |
| `token` | token id / edition / mint |
| `side` | `sell` or `buy` |
| `price` | integer amount in the payment asset's base units |
| `pay` | payment asset or currency identifier |
| `expiry` | absolute time or source-chain height after which the signal is dead |
| `maker` | the address that can settle it |
| `nonce` | replay protection |
| `adapter` | settlement adapter and version |
| `ref` | hash of the full terms, including any SporeM3 payload |

The public record should not contain a buyer's private identity, a shipping address, or an unencrypted private
listing. Those belong in the SporeM3 payload or off-chain.

## DERO signal board

- A small, append-only contract or an existing social/chat surface can hold the public records. Selection is an
  implementation decision; no new contract is committed by this note.
- Writes cost DERO gas. The board has no custody, no owner-controlled delisting, and no ability to settle.
- Readers index by chain/asset/token and expiry. The app verifies the source chain before showing an actionable
  trade; a signal alone only earns a "listed elsewhere" label.
- Cancellation and expiry are explicit signals. A disappeared maker is not a settlement.

## SporeM3 fallback

Use SporeM3 when the public DERO board is unavailable, or when the terms must be private:

- E2E-encrypted messages over DERO transaction payloads, with the repo's 144-byte payload limit requiring
  chunking or a hash-plus-out-of-band-document pattern.
- The message carries the full terms and the same `ref` hash as the public signal.
- If SporeM3 is unavailable, the trade waits; it never falls back to a custodial relay or a server.
- Spore metadata is linkable. Do not describe it as anonymous; expose the timing/linkability cost.

## Settlement adapters

- **Ethereum:** verify the collection's `ownerOf`/approval and the marketplace order on the source chain; settle
  with the wallet's normal EIP-712/Seaport flow. Royalties and fees are Ethereum-side facts, not DERO signals.
- **Solana:** verify the mint/token account and the program's authority; settle with a Solana transaction.
- **DERO:** optional receipt/attestation only. No wrapped NFT and no foreign-value custody.
- **wXMR/wZEC:** parked. Existing contracts and tests stay in the tree, but they are not the launch dependency
  for this direction and no new work should assume they are.

## Threat model

| Risk | Control |
|---|---|
| Fake listing | The taker verifies ownership and approval on the source chain before signing. A signal is never proof. |
| Signal replay after a sale | Nonce, expiry, and a source-chain ownership check at execution time. |
| Front-running / stale price | Short expiry, explicit price, and a source-chain read immediately before settlement. |
| Censorship | Public DERO records; SporeM3 fallback; no operator-run relay required. |
| Signal spam | DERO postage and per-maker rate/expiry rules; no free unbounded board. |
| Private-data leak | Public record carries only `ref`; private terms stay in the SporeM3 payload. |
| False receipt | Receipts are hashes of source-chain settlement; the app links the source transaction; a receipt is not a transferable claim. |

## Work order

1. **Schema and validator.** Define the signal record, canonical serialization, expiry/nonce rules, and tests.
   No chain calls.
2. **Read-only DERO board.** Publish/read/cancel with the minimum fields; index by chain/asset/token. No custody.
3. **Ethereum adapter.** Read ownership/approval, verify the maker's order, then settle through the wallet.
4. **Solana adapter.** The same verification/settlement split for Solana NFTs.
5. **SporeM3 transport.** Private/fallback path using the same `ref`; document payload limits and linkability.
6. **Park / re-scope the vault track.** Keep the existing wXMR/wZEC code and evidence; stop treating it as the
   path to NFT liquidity.

## Open decisions

- Public-by-default signals, with SporeM3 only for private terms (proposed), or private-by-default?
- Signal postage paid in DERO, or sponsored by the maker's settlement adapter?
- A new minimal signal contract, or an existing `chat.bas`/`rate.bas` surface?
- Which Ethereum marketplace/order standard is the first adapter, and which Solana program?
- Whether a DERO receipt is worth writing at all, given it adds a second chain to every trade.

## Relationship to the existing docs

This note changes the priority, not the history. `DECISIONS.md` §10.15 records the ruling. The ZEC/XMR-first
language in `docs/LOOP_PROMPT.md` and `docs/BACKLOG.md` is marked as legacy for new work; the existing vault
contracts, tests and evidence remain. Nothing here claims DERO custody, a bridge, or a trustless foreign-NFT
verification.
