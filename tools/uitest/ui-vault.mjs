// WT-17's browser pass: the Wrap screen must show the connected wallet's index, its attested deposit
// address and its withdrawal #j rows, and the Burn probe must match the contract's own record. It is
// read-only: it sets the page's wallet address to the owner of the configured vault's index 0 (read from
// the contract) and never signs anything. Run by ui.mjs in the ordinary simulator pass.
//
//   node tools/uitest/ui.mjs http://127.0.0.1:8082/index.html .sim/ui-ro
//
// The fixture is the simulator's configured XMR vault, which tools/simvault installs and the signers
// attest. A build with no vault, or a simulator whose index 0 has not been attested, fails here instead
// of silently passing: that is the state in which the Wrap screen cannot report a Register or a Burn.
export async function vaultPass({ browser, url, out, step, fail, sleep }) {
  const p = await browser.newPage({ viewport: { width: 430, height: 900 }, bypassCSP: true });
  const errs = [];
  p.on("pageerror", (e) => errs.push("pageerror: " + e.message));
  p.on("console", (m) => { if (m.type() === "error" && !/net::ERR_|favicon/.test(m.text())) errs.push("console: " + m.text()); });
  await p.goto(url, { waitUntil: "load" });
  await p.locator("#introOk").click({ timeout: 8000 }).catch(() => {});
  await p.waitForFunction(() => window.Veil && window.Veil.data && window.Veil.data.lastRefresh > 0, null, { timeout: 60000 }).catch(() => {});
  const txt = async () => (await p.locator("#screen").innerText()).replace(/\s+/g, " ");
  const shot = (n) => p.screenshot({ path: `${out}/wt17-${n}.png`, fullPage: true });

  const fixture = await p.evaluate(async () => {
    const D = window.Veil.data, V = window.Veil, rec = D.vaults().find((v) => /xmr/i.test(String(v.sym || ""))) || D.vaults()[0];
    if (!rec) return { skip: "no configured vault" };
    const scid = rec.scid;
    const r = await V.node.getSC(scid, { keys: ["ua_0", "da_0", "dh_0", "dver_0", "wn"] });
    const owner = /^[0-9a-f]+$/.test(r.vals.ua_0 || "") ? D.rawToAddr(r.vals.ua_0) : "";
    if (!owner) return { err: "the configured vault has no index 0" };
    const addr = /^[0-9a-f]+$/.test(r.vals.da_0 || "") ? V.hexToStr(r.vals.da_0) : "";
    if (!addr) return { err: "index 0 is not attested on the configured vault", owner };
    // The screen reads myVault only for a connected wallet. The address is enough for this read-only pass:
    // no transfer is built, and nothing is signed.
    V.wallet.canon = owner;
    V.wallet.connected = true;
    const w = await D.loadVault(scid);
    const my = await D.myVault(owner);
    return { scid, owner, addr, w: !!w, indexes: my && my.indexes,
      withdrawals: ((my && my.withdrawals) || []).map((x) => ({ ...x, units: x.units.toString(), fee: x.fee.toString() })) };
  });
  if (fixture.skip) { step("WT-17: " + fixture.skip + "; the Wrap-screen pass was not exercised"); await p.close(); return; }
  if (fixture.err) { fail("WT-17 fixture: " + fixture.err); await p.close(); return; }
  step(`WT-17 fixture: ${fixture.scid.slice(0, 10)}… index 0 owner ${fixture.owner.slice(0, 14)}… address ${fixture.addr.slice(0, 14)}…`);

  await p.evaluate((s) => { location.hash = "#wrap/" + s; }, fixture.scid);
  await p.waitForFunction(() => /Limits and fees|could not be read/i.test(document.getElementById("screen").innerText), null, { timeout: 45000 }).catch(() => {});
  await sleep(800);
  await shot("wrap");
  const t = await txt();
  if (!/Live since block \d+/.test(t)) fail("WT-17: the Wrap screen does not render the attested deposit address: " + t.slice(0, 300));
  if (!t.includes(fixture.addr)) fail("WT-17: the Wrap screen does not show the attested address " + fixture.addr.slice(0, 20) + "…: " + t.slice(0, 300));
  if (!/Register another index|Register a deposit index/.test(t)) fail("WT-17: the Wrap screen does not show the registered index action: " + t.slice(0, 300));
  step("WT-17: index and deposit address render");

  const x = fixture.withdrawals[0];
  if (!x) {
    step("WT-17: no withdrawal for index 0 on this simulator; the #j row and Burn probe were not exercised");
  } else {
    if (!t.includes("#" + x.j)) fail("WT-17: withdrawal #" + x.j + " has no #j row on the Wrap screen: " + t.slice(0, 400));
    if (x.dest && !t.includes(x.dest)) fail("WT-17: withdrawal #" + x.j + "'s destination is not rendered");
    const probe = await p.evaluate(async (f) => {
      const D = window.Veil.data, w = D.vault, x = f.x;
      const units = BigInt(x.units), fee = BigInt(x.fee);
      const good = await D.burnOpened(w, f.owner, x.j, x.j + 1, x.dest, units + fee);
      const wrong = await D.burnOpened(w, f.owner, x.j, x.j + 1, x.dest, units + fee + 1n);
      return { good, wrong };
    }, { owner: fixture.owner, x });
    if (!probe.good) fail("WT-17: the Burn probe did not find withdrawal #" + x.j + " by its own record");
    if (probe.wrong) fail("WT-17: the Burn probe matched a different net amount, so it cannot prove our Burn executed");
    step(`WT-17: withdrawal #${x.j} row renders; the Burn probe matches the record and rejects a wrong net amount`);
  }

  await p.evaluate(() => { window.Veil.wallet.connected = false; window.Veil.wallet.canon = null; });
  if (errs.length) fail("WT-17 page errors: " + errs.join("; "));
  await p.close();
}
