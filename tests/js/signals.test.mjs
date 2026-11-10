// The first read-only signal board is a local reader over app/signal.js. These tests exercise its pure board
// parser and expiry status without a DOM or a chain.
import { test } from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { join } from "node:path";
import { ROOT } from "./harness.mjs";

function load() {
  const window = { Veil: {} };
  new Function("window", readFileSync(join(ROOT, "app", "signal.js"), "utf8"))(window);
  const noop = () => {};
  window.Veil.ui = { h: noop, render: noop, button: noop, router: { on: noop }, emptyState: noop, clean: (s, max = 64) => String(s || "").slice(0, max) };
  new Function("window", readFileSync(join(ROOT, "app", "signals.js"), "utf8"))(window);
  return window.Veil;
}

const V = load();
const X = V.signals;

test("the local board's example has two accepted records and one field rejection", () => {
  const cases = X.decode(X.examples);
  assert.equal(cases.length, 3);
  assert.deepEqual(cases.map((c) => c.ok), [true, true, false]);
  assert.equal(cases[0].rec.chain, "eth");
  assert.equal(cases[1].rec.chain, "sol");
  assert.equal(cases[2].at, 3);
  assert.equal(cases[2].field, "token");
  assert.match(cases[2].message, /canonical/);
  assert.ok(!cases[2].message.startsWith("token: "), "the field is shown separately, not duplicated in the message");
});

test("the board reads every valid shared vector and keeps canonical bytes", () => {
  const vectors = JSON.parse(readFileSync(join(ROOT, "tools", "signal", "testdata", "signal_v1_vectors.json"), "utf8"));
  const records = vectors.records.filter((v) => v.valid);
  const cases = X.decode(records.map((v) => JSON.stringify(v.record)).join("\n"));
  assert.equal(cases.length, records.length);
  cases.forEach((c, i) => {
    assert.equal(c.ok, true, records[i].name);
    assert.equal(c.canonical, records[i].canonical, records[i].name);
  });
});

test("a malformed line keeps its own line number and rejection field", () => {
  const cases = X.decode('\n{"v":1}\nnot json\n');
  assert.equal(cases.length, 2);
  assert.equal(cases[0].at, 2);
  assert.equal(cases[0].field, "chain");
  assert.equal(cases[1].at, 3);
  assert.equal(cases[1].field, "signal");
});

test("time and height expiry are reported without inventing a source height", () => {
  const cases = X.decode(X.examples);
  assert.equal(X.status(cases[0], 1).label, "Live");
  assert.equal(X.status(cases[0], 2000000000).label, "Expired");
  assert.equal(X.status(cases[1], 0).label, "Height expiry");
  assert.match(X.status(cases[1], 0).note, /source height not checked/);
  assert.equal(X.status(cases[2], 0), null);
});
