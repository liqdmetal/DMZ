// Cross-language differential for the v1 signal record. tools/signal and app/signal.js read the same
// fixture, so the browser and the tools cannot accept different records. Regenerate the fixture only
// by changing the rules in both implementations, never by editing one side's expected result.
import { test } from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { join } from "node:path";
import { ROOT } from "./harness.mjs";

function loadSignal() {
  const window = {};
  new Function("window", readFileSync(join(ROOT, "app", "signal.js"), "utf8"))(window);
  return window.Veil.signal;
}

const S = loadSignal();
const VECTORS = JSON.parse(readFileSync(join(ROOT, "tools", "signal", "testdata", "signal_v1_vectors.json"), "utf8"));

function rejection(fn) {
  try { fn(); } catch (e) { return e; }
  return null;
}

function rejected(fn, field, label) {
  const e = rejection(fn);
  assert.ok(e, label + " was accepted");
  if (field) assert.equal(e.field, field, label + ": " + e.message);
  return e;
}

test("the browser mirror reads the shared v1 fixture", () => {
  assert.equal(S.version, VECTORS.version);
  assert.deepEqual([...S.fields], ["v", "chain", "asset", "token", "side", "price", "pay", "expiry", "maker", "nonce", "adapter", "ref"]);
});

test("the browser build loads signal.js before the screen code", () => {
  const html = readFileSync(join(ROOT, "app", "index.html"), "utf8");
  const at = html.indexOf('src="signal.js"');
  assert.ok(at > 0, "signal.js is loaded");
  assert.ok(at < html.indexOf('src="app.js"'), "signal.js loads before app.js");
});

test("every shared record has the same accept/reject result and canonical bytes", () => {
  for (const tc of VECTORS.records) {
    const raw = JSON.stringify(tc.record);
    if (!tc.valid) {
      rejected(() => S.validate(tc.record), tc.field, tc.name + " (validate)");
      rejected(() => S.parse(raw), tc.field, tc.name);
      continue;
    }
    const got = S.parse(raw);
    assert.equal(S.validate(got), true, tc.name + " validate");
    assert.equal(S.canonical(got), tc.canonical, tc.name + " canonical bytes");
    for (const lc of tc.live || []) {
      if (lc.valid) {
        assert.equal(S.live(got, { now: lc.now, height: lc.height }), true, tc.name + " live at now=" + lc.now + " height=" + lc.height);
      } else {
        rejected(() => S.live(got, { now: lc.now, height: lc.height }), lc.field, tc.name + " live at now=" + lc.now + " height=" + lc.height);
      }
    }
  }
});

test("the browser parser rejects the same non-record JSON", () => {
  for (const tc of VECTORS.json) {
    if (tc.valid) {
      assert.doesNotThrow(() => S.parse(tc.raw), tc.name);
    } else {
      rejected(() => S.parse(tc.raw), tc.field, tc.name);
    }
  }
});

test("the browser expiry parser matches the Go parser", () => {
  for (const tc of VECTORS.expiry) {
    if (!tc.valid) {
      rejected(() => S.parseExpiry(tc.input), "", "expiry " + tc.input);
      continue;
    }
    const got = S.parseExpiry(tc.input);
    assert.equal(got.kind, tc.kind, "expiry kind for " + tc.input);
    assert.equal(got.value.toString(), tc.value, "expiry value for " + tc.input);
  }
});

test("canonical bytes ignore input key order and whitespace", () => {
  const eth = VECTORS.records.find((r) => r.name === "ethereum sell");
  const shuffled = JSON.stringify(Object.fromEntries(Object.entries(eth.record).reverse()), null, 2);
  assert.equal(S.canonical(S.parse(shuffled)), eth.canonical);
});

test("duplicate keys are rejected at any depth, not merged by JSON.parse", () => {
  const base = JSON.stringify(VECTORS.records.find((r) => r.name === "ethereum sell").record);
  rejected(() => S.parse(base.replace('"v":1', '"v":1,"v":1')), "signal", "top-level duplicate");
  rejected(() => S.parse(base.slice(0, -1) + ',"extra":{"a":1,"a":2}}'), "signal", "nested duplicate");
});
