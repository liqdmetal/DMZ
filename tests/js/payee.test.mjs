// app/rpc.js: what a contract payout can actually pay (AUD-15-04). A …To entrypoint pays ADDRESS_RAW(dest) —
// NewAddress(dest).Compressed(), the bare 33-byte key — and SEND_DERO_TO_ADDRESS / SEND_ASSET_TO_ADDRESS carry no
// payload, so an integrated address's payment id and port never reach the receiver. payKey therefore names a payee
// only when there is nothing to drop, and app.js/token.js/nfas.js refuse everything else instead of signing a call
// whose proceeds would arrive unattributable at an exchange.
//
// The integrated vectors are the shape rpc/address.go writes (rpc/address.go MarshalText: version byte 1, the
// 33-byte point, then rpc.Arguments.MarshalBinary, a msgpack map, under the deroi/detoi hrp), built from the real
// mainnet and testnet addresses below with the same bech32 the app implements. The trailing bytes are arbitrary:
// what this file pins is that any argument bytes are refused whole.
//
//   node --test tests/js
import test from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { join } from "node:path";
import { ROOT } from "./harness.mjs";

const window = { Veil: {} };
new Function("window", readFileSync(join(ROOT, "app", "rpc.js"), "utf8"))(window);
const V = window.Veil;

const DERO = "dero1qy0khp9s9yw2h0eu20xmy9lth3zp5cacmx3rwt6k45l568d2mmcf6qgcsevzx";
const DETO = "deto1qy2nxgts7wdn28ckc4l2tewphjcppqjfj69ddkxjn0ay8hlsjx73jqgmat5s8";
// the same keys as an exchange or wallet prints them when they carry a payment id
const DEROI = "deroi1qy0khp9s9yw2h0eu20xmy9lth3zp5cacmx3rwt6k45l568d2mmcf6qg3yge5g4txw7yqsdta8p";
const DETOI = "detoi1qy2nxgts7wdn28ckc4l2tewphjcppqjfj69ddkxjn0ay8hlsjx73jqg3yge5g4txw7yqdmrt2j";
// the testnet key above, spelled as the DVM writes it (ADDRESS_STRING is always the mainnet hrp: the key is what
// the contract pays, and UnmarshalText does not check the network)
const DETO_KEY = "dero1qy2nxgts7wdn28ckc4l2tewphjcppqjfj69ddkxjn0ay8hlsjx73jqgpyhv40";
// an integrated hrp with no argument bytes: rpc/address.go hands the empty tail to Arguments.UnmarshalBinary,
// which fails, so the wallet refuses it too
const DEROI_BARE = "deroi1qy0khp9s9yw2h0eu20xmy9lth3zp5cacmx3rwt6k45l568d2mmcf6qgkfqdx3";
// a plain hrp with trailing words: rpc/address.go's default branch, "invalid address length as per spec"
const DERO_TRAILING = "dero1qy0khp9s9yw2h0eu20xmy9lth3zp5cacmx3rwt6k45l568d2mmcf6qdte5ynrzxd";

test("a plain dero1/deto1 is payable, and the key it pays is what the DVM writes", () => {
  assert.equal(V.payKey(DERO), DERO);
  assert.equal(V.isPayable(DERO), true);
  assert.equal(V.addrParts(DERO).extra, 0);
  // a testnet address is paid by key, and the key is what the DVM stores
  assert.equal(V.payKey(DETO), DETO_KEY);
  assert.equal(V.isPayable(DETO), true);
});

test("an integrated address is refused, and canonAddress still names the key it would have paid", () => {
  for (const [a, key] of [[DEROI, DERO], [DETOI, DETO_KEY]]) {
    assert.equal(V.payKey(a), null, a);
    assert.equal(V.isPayable(a), false, a);
    assert.ok(V.addrParts(a).extra > 0, "arguments are present");
    // what the old build did silently: strip the arguments and pay the bare key
    assert.equal(V.canonAddress(a), key);
    // and the payment id is not recoverable from what canonAddress returns
    assert.ok(!/^deroi/.test(V.canonAddress(a)));
  }
  // a …To call would store this raw key, with no payload to attribute the funds to
  assert.equal(V.addressRaw(DEROI), V.addressRaw(DERO));
});

test("every address the wallet itself refuses is refused here", () => {
  assert.equal(V.payKey(DEROI_BARE), null, "integrated hrp, no arguments to encode");
  assert.equal(V.payKey(DERO_TRAILING), null, "plain hrp with trailing words");
  // canonAddress stays the lenient parse other reads rely on (feeds compare keys, not arguments)
  assert.equal(V.canonAddress(DEROI_BARE), DERO);
  assert.equal(V.canonAddress(DERO_TRAILING), DERO);
});

test("junk is null, never a throw", () => {
  for (const a of ["", "dero", "dero1", "x".repeat(64), "dero1qy0", null, undefined, 42, {}, DERO.toUpperCase().slice(0, -1)])
    assert.equal(V.payKey(a), null, String(a));
  // bech32 is case-insensitive as a whole string, never mixed
  assert.equal(V.payKey(DERO.toUpperCase()), DERO);
});

// AUD-15-15: a refused private trade never reverts. refundto pays everything attached to dest and stores
// "rf_" + HEX(TXID()) (contracts/launch.bas:194, pool.bas:216); the app reads that marker back by key, so the key
// has to be the one the DVM writes.
test("the refusal marker is the key the contract writes, whatever case the id arrives in", () => {
  const txid = "a1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f90";
  assert.equal(V.refundKey(txid), "rf_" + txid);
  // a wallet or node that spells the transaction id in upper case still names the same marker: HEX in the DVM is lower
  assert.equal(V.refundKey(txid.toUpperCase()), "rf_" + txid, "a refusal must not read as no marker because of case");
  // anything that is not a 32-byte id names no key at all, rather than a key that silently misses
  for (const bad of ["", "0x" + txid, txid.slice(0, -2), txid + "00", null, undefined, 42, {}])
    assert.equal(V.refundKey(bad), null, String(bad));
});

test("the round trip the contracts rely on: payKey → addressRaw → rawToAddress", () => {
  const raw = V.addressRaw(V.payKey(DERO));
  assert.match(raw, /^[0-9a-f]{66}$/);
  assert.equal(V.rawToAddress(raw), DERO);
  // an integrated address has no such round trip, which is why it is refused rather than passed through
  assert.equal(V.addressRaw(V.payKey(DEROI)), null);
});
