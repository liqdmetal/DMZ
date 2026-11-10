// Randomized differential corpus for v1 signals. tools/signal/random_test.go generates the fixture from
// a fixed seed: valid records on both chains plus every targeted mutation. This test runs the same
// corpus through app/signal.js, including the canonical bytes, live checks and rejection field.
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
const RANDOM = JSON.parse(readFileSync(join(ROOT, "tools", "signal", "testdata", "signal_v1_random_vectors.json"), "utf8"));

function rejected(fn, field, label) {
  let err;
  try { fn(); } catch (e) { err = e; }
  assert.ok(err, label + " was accepted");
  if (field) assert.equal(err.field, field, label + ": " + err.message);
}

test("the randomized corpus is broad and reproducible", () => {
  assert.equal(RANDOM.version, S.version);
  assert.equal(RANDOM.seed, 0x51a1);
  assert.equal(RANDOM.rounds, RANDOM.cases.length);
  const valid = RANDOM.cases.filter((c) => c.valid).length;
  const invalid = RANDOM.cases.length - valid;
  assert.ok(valid >= 10, "at least 10 valid randomized records");
  assert.ok(invalid >= 100, "at least 100 mutated records");
});

test("app/signal.js agrees with every randomized case", () => {
  for (const tc of RANDOM.cases) {
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
