// DMZ signals.js — the first read-only signal board. There is no signal-board contract in this build yet,
// so this screen reads a local board (one JSON record per line, pasted or opened from a file) and runs every
// line through app/signal.js, the browser mirror of tools/signal. It never fetches a chain, connects a wallet,
// signs, publishes, or stores anything. A valid record is discovery data, not proof of ownership or settlement.
(function () {
  const V = window.Veil, U = V.ui, S = V.signal;
  const { h, render, button, router, emptyState } = U;

  // A board line is one v1 record. The examples are the schema's own fixtures plus one record whose token id
  // is deliberately not canonical, so the screen can show the field-level rejection path without a chain.
  const ETH = {
    v: 1, chain: "eth", asset: "0x1111111111111111111111111111111111111111", token: "0",
    side: "sell", price: "1000000000000000000", pay: "0x3333333333333333333333333333333333333333",
    expiry: "time:2000000000", maker: "0x2222222222222222222222222222222222222222",
    nonce: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    adapter: "seaport/1.6.0", ref: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
  };
  const SOL = {
    v: 1, chain: "sol", asset: "So11111111111111111111111111111111111111112",
    token: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", side: "buy", price: "2500000", pay: "native",
    expiry: "height:250000000", maker: "9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM",
    nonce: "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
    adapter: "magiceden/2.0.0", ref: "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
  };
  const EXAMPLES = [JSON.stringify(ETH), JSON.stringify(SOL), JSON.stringify(Object.assign({}, ETH, { token: "01" }))].join("\n");

  const messageOf = (e, field) => {
    const text = String((e && e.message) || e || "rejected");
    const prefix = field + ": ";
    return U.clean(text.startsWith(prefix) ? text.slice(prefix.length) : text, 260);
  };

  // Strict per-line parsing, so a duplicate key, trailing data or a non-canonical value is rejected by the same
  // validator the Go tool mirrors. Blank lines are skipped; `at` is the line number the record came from.
  function decode(text) {
    const out = [];
    String(text || "").split(/\r?\n/).forEach((line, i) => {
      const raw = line.trim();
      if (!raw) return;
      try {
        const rec = S.parse(raw);
        out.push({ ok: true, at: i + 1, raw, rec, expiry: S.parseExpiry(rec.expiry), canonical: S.canonical(rec) });
      } catch (e) {
        const field = (e && e.field) || "signal";
        out.push({ ok: false, at: i + 1, raw, field, message: messageOf(e, field) });
      }
    });
    return out;
  }

  // Validation says the record is well formed; Live says a time expiry has passed. A height expiry is never
  // guessed: without the source chain's height the screen says so instead of calling the record live.
  function status(item, now) {
    if (!item || !item.ok) return null;
    const e = item.expiry;
    if (e.kind === "height") return { label: "Height expiry", cls: "dim", note: "source height not checked" };
    const seconds = e.value;
    const date = seconds <= 8640000000000n ? new Date(Number(seconds) * 1000).toISOString().replace(".000Z", "Z") : String(seconds);
    const clock = typeof now === "bigint" ? now : BigInt(Math.floor(now === undefined ? Date.now() / 1000 : now));
    const dead = clock >= seconds;
    return { label: dead ? "Expired" : "Live", cls: dead ? "bad" : "up", note: "time " + date };
  }

  const unknown = (f) => f !== "signal" && !S.fields.includes(f);

  function rawDetails(raw) {
    return h("details", { class: "more" }, h("summary", null, "The line this was read from"),
      h("div", { class: "note mono mt8", style: "white-space:pre-wrap;overflow-wrap:anywhere" }, raw));
  }

  function rejected(item) {
    return h("article", { class: "card mt", "data-signal-rejected": item.at, "data-field": item.field },
      h("div", { class: "between" },
        h("b", null, "Rejected record " + item.at),
        h("span", { class: "pill bad mono" }, item.field)),
      h("div", { class: "sub mt8" },
        "The validator stopped at ", h("span", { class: "mono" }, item.field), ": ", item.message),
      unknown(item.field) ? h("div", { class: "note mt8" }, "This field is not part of signal v1, so the rest of the record was not read.") : null,
      rawDetails(item.raw));
  }

  function accepted(item) {
    const r = item.rec, st = status(item);
    const row = (label, value) => h("div", { class: "kv" },
      h("span", { class: "dim" }, label),
      h("b", { class: "mono" }, value));
    return h("article", { class: "card mt", "data-signal-valid": item.at },
      h("div", { class: "between" },
        h("div", { class: "row tight" },
          h("span", { class: "pill" }, r.chain === "eth" ? "Ethereum" : "Solana"),
          h("span", { class: "pill" }, r.side),
          st ? h("span", { class: "pill " + st.cls }, st.label) : null),
        h("span", { class: "mono tiny dim2" }, "line " + item.at)),
      h("div", { class: "stack mt8" },
        row("Asset", r.asset),
        row("Token / mint", r.token),
        row("Price", r.price),
        row("Pay", r.pay),
        row("Maker", r.maker),
        row("Adapter", r.adapter),
        row("Expiry", r.expiry + (st && st.note ? " · " + st.note : ""))),
      h("details", { class: "more" }, h("summary", null, "Canonical bytes"),
        h("div", { class: "note mono mt8", style: "overflow-wrap:anywhere" }, item.canonical)));
  }

  let lastText = "", lastCases = [];

  router.on("signals", (el) => {
    const input = h("textarea", {
      class: "mono", rows: "6", spellcheck: "false", autocomplete: "off",
      placeholder: '{"v":1,"chain":"eth",…}', "aria-label": "Signal records, one JSON record per line", "aria-describedby": "sigHelp",
      style: "width:100%;min-height:132px;resize:vertical",
    });
    const file = h("input", { type: "file", accept: ".json,.jsonl,text/plain,application/json", style: "display:none" });
    const summary = h("div", { class: "sub mt", role: "status", "aria-live": "polite" });
    const results = h("div");
    const paint = () => {
      const ok = lastCases.filter((c) => c.ok).length, bad = lastCases.length - ok;
      render(summary, lastCases.length
        ? ok + " accepted · " + bad + " rejected · read locally, not from a chain"
        : "Nothing read yet. A signal is one JSON object on one line.");
      if (!lastCases.length) {
        render(results, emptyState("feed", "No records read",
          "Paste a v1 signal below, one JSON record per line, or open a board file. This first screen is read-only: it does not fetch a chain, publish, sign or store anything.",
          button("Read the examples", "small primary", () => read(EXAMPLES))));
      } else render(results, lastCases.map((c) => (c.ok ? accepted(c) : rejected(c))));
    };
    const read = (text) => {
      lastText = String(text || "");
      input.value = lastText;
      lastCases = decode(lastText);
      paint();
    };
    file.onchange = async () => {
      const f = file.files && file.files[0];
      if (!f) return;
      try { read(await f.text()); }
      catch (e) { lastCases = [{ ok: false, at: 1, raw: "", field: "signal", message: "the file could not be read" }]; paint(); }
      file.value = "";
    };
    input.value = lastText;
    paint();
    render(el,
      h("div", { class: "between mt" },
        h("h1", null, "Signal board"),
        h("span", { class: "pill dim" }, "Read-only")),
      h("div", { class: "sub mb" },
        "Public discovery intents checked against signal v1. The same validator runs in the Go tool and in this page; a record that fails is shown with the field that failed."),
      h("div", { class: "warnbox mb" },
        h("b", null, "Validation is not verification. "),
        "A valid signal says only that the record is well formed. It is not proof that the maker owns the NFT, approved a marketplace, or can settle, and it is not a price or a promise. Check ownership and settlement on the NFT's native chain."),
      h("div", { class: "note mb" },
        "There is no signal-board contract in this build yet, so this first screen reads one JSON record per line from you. Nothing is fetched, published, signed or stored."),
      h("div", { class: "card" },
        h("div", { class: "between" },
          h("h2", null, "Read a board"),
          h("div", { class: "row tight" },
            button("Open a file", "small ghost", () => file.click()),
            button("Load examples", "small ghost", () => read(EXAMPLES)),
            button("Clear", "small ghost", () => { lastText = ""; lastCases = []; input.value = ""; paint(); input.focus(); }))),
        h("label", { class: "field mt" }, "One signal per line · v1 JSON", input),
        h("div", { class: "note mt8", id: "sigHelp" }, "Blank lines are ignored. Every other line is parsed strictly: duplicate keys, unknown fields, trailing data and non-canonical values are refused."),
        h("div", { class: "row mt8" },
          button("Read board", "small primary", () => read(input.value)),
          file)),
      summary,
      results);
  });

  V.signals = { decode, status, examples: EXAMPLES };
})();
