// WT-69: the app must not send a height-gated call through DERO.GetGasEstimate. The daemon's estimator runs the
// contract at BLOCK_HEIGHT() = 0, so Reclaim (past rel + 20000) and Register (after a lapsed halt) are always
// refused there, and app.tx turns that refusal into "would fail" without ever asking the wallet. These tests drive
// the real app.tx with a node whose estimator refuses the call at height 0 and a wallet that would accept it.
//
//   node --test tests/js/wrap.heightgate.test.mjs
import test from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { join } from "node:path";
import { ROOT } from "./harness.mjs";

const SCID = "e".repeat(64);
const ADDR = "deto1qyyvq8xp22kkzta3g25r8eful0nqsm9jr23522mqfn4xmp89skhkgqqc62nc9";
const src = (n) => readFileSync(join(ROOT, "app", n), "utf8");

// Load app/app.js the way the browser does, with only the collaborators app.tx reaches. The estimator always
// refuses with the height-gate error the real daemon produces; the wallet would accept the transaction.
function loadApp() {
  const toasts = [], transfers = [], estimates = [];
  const window = { Veil: {}, addEventListener() {} };
  const V = window.Veil;
  V.ui = {
    h: (tag, props, ...kids) => ({ tag, props, kids }),
    toast: (msg, kind) => toasts.push({ msg, kind }),
    openSheet() {}, closeSheet() {}, button() {}, kv() {}, icon() {},
    router: { dispatch() {}, go() {}, current: null, on() {} },
    $: () => null, $$: () => [], clean: (s, n) => String(s || "").slice(0, n),
    render() {}, timeAgo() {}, details() {},
  };
  V.amm = { fmtUnits: (x) => String(x), fmtDero: (x) => String(x) };
  V.data = {
    cfg: { network: "simulator" }, assets: new Map(), nfas: new Map(), nfts: new Map(),
    onChange() {}, lastRefresh: 0, lastError: "",
  };
  V.wallet = {
    connected: true, address: ADDR,
    call: async (m) => (m === "DERO.GetInfo" ? { network: "simulator" } : {}),
    transfer: async (p) => { transfers.push(p); return { txid: "a".repeat(64) }; },
    on() {}, getBalance: async () => ({ balance: "0" }),
  };
  V.live = { onBlock() {}, start() {} };
  V.stats = {};
  V.allow = { check: async () => null };
  V.bech32Decode = (a) => (String(a).startsWith("deto1") ? { hrp: "deto" } : { hrp: "dero" });
  V.endpointURL = () => ({ url: "", loopback: true });
  V.toBig = (x) => BigInt(x || 0);
  V.shortAddr = (a) => a;
  V.node = {
    setEndpoint() {},
    call: async (m, p) => {
      if (m === "DERO.GetGasEstimate") { estimates.push(p); throw new Error("contract refused at BLOCK_HEIGHT() 0: height gate"); }
      return {};
    },
    info: async () => ({ network: "simulator", height: 1 }),
    randomAddress: async () => ADDR,
    getTx: async () => ({ valid_block: true, in_pool: false, block_height: 1 }),
    getSC: async () => ({ vals: {}, raws: {} }),
  };
  const store = new Map();
  const localStorage = { getItem: (k) => (store.has(k) ? store.get(k) : null), setItem: (k, v) => store.set(k, v), removeItem: (k) => store.delete(k) };
  const document = {
    hidden: false, documentElement: { classList: { add() {} }, dataset: {} },
    getElementById: () => null, querySelector: () => null, addEventListener() {},
  };
  new Function("window", "document", "localStorage", "sessionStorage", "addEventListener", src("app.js"))(
    window, document, localStorage, { ...localStorage }, () => {},
  );
  V.app.scheduleRefresh = () => {};
  V.app.waitTx = async () => ({ state: "included", height: 1 });
  return { V, toasts, transfers, estimates };
}

for (const [entry, args] of [["Reclaim", [["j", "U", 0n]]], ["Register", []]]) {
  test(`WT-69: ${entry} is sent at a fixed fee, not refused by the height-0 estimator`, async () => {
    const { V, transfers, estimates } = loadApp();
    const txid = await V.app.tx({ scid: SCID, entry, args, describe: entry });
    assert.equal(estimates.length, 0, `${entry} must not be estimated: DERO.GetGasEstimate runs the contract at BLOCK_HEIGHT() 0`);
    assert.equal(transfers.length, 1, `${entry} must reach the wallet, not stop at the estimator`);
    assert.ok(Number.isInteger(transfers[0].fees) && transfers[0].fees > 0 && transfers[0].fees <= 100000,
      `${entry} must use a positive fee within the app's 1 DERO cap`);
    assert.equal(txid, "a".repeat(64), `${entry} must return the wallet's txid`);
  });
}
