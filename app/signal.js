// DMZ signal.js — browser mirror of tools/signal. It accepts exactly the public DERO signal v1 records
// that the Go validator accepts, and returns the same field on every rejection. It performs no chain
// calls and treats a valid signal only as well-formed discovery data: ownership, approval and settlement
// still have to be checked on the NFT's native chain.
(function () {
  const V = (window.Veil = window.Veil || {});
  const VERSION = 1;
  const FIELDS = Object.freeze(["v", "chain", "asset", "token", "side", "price", "pay", "expiry", "maker", "nonce", "adapter", "ref"]);
  const MAX_U64 = (1n << 64n) - 1n;
  const MAX_I64 = (1n << 63n) - 1n;
  const BASE58 = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz";
  const ADAPTER = /^[a-z][a-z0-9-]{0,31}\/[0-9]+\.[0-9]+\.[0-9]+$/;
  const CANONICAL_UINT = /^(0|[1-9][0-9]*)$/;
  const LOWERCASE_HEX = /^[0-9a-f]+$/;

  const err = (field, message) => {
    const e = new Error(field + ": " + message);
    e.field = field;
    return e;
  };

  function validate(rec) {
    if (rec === null) throw err("v", "got 0, this reader accepts " + VERSION);
    if (!rec || typeof rec !== "object" || Array.isArray(rec)) throw err("signal", "record is not an object");
    for (const key of Object.keys(rec)) if (!FIELDS.includes(key)) throw err(key, "unknown field");
    if (!Number.isInteger(rec.v) || rec.v !== VERSION) throw err("v", "got " + String(rec.v) + ", this reader accepts " + VERSION);
    if (rec.chain !== "eth" && rec.chain !== "sol") throw err("chain", JSON.stringify(rec.chain) + " is not eth or sol");
    for (const f of FIELDS) {
      if (f === "v") continue;
      if (typeof rec[f] !== "string") throw err(f, "must be a string");
    }
    if (!validAddress(rec.chain, rec.asset)) throw err("asset", JSON.stringify(rec.asset) + " is not a canonical " + rec.chain + " address");
    if (!validToken(rec.chain, rec.token)) throw err("token", JSON.stringify(rec.token) + " is not a canonical " + rec.chain + " token id");
    if (rec.side !== "sell" && rec.side !== "buy") throw err("side", JSON.stringify(rec.side) + " is not sell or buy");
    if (!validUint(rec.price, 256, false)) throw err("price", JSON.stringify(rec.price) + " is not a positive canonical uint256");
    if (rec.pay !== "native" && !validAddress(rec.chain, rec.pay)) throw err("pay", JSON.stringify(rec.pay) + " is neither native nor a canonical " + rec.chain + " address");
    try { parseExpiry(rec.expiry); } catch (e) { throw err("expiry", e.message.replace(/^expiry: /, "")); }
    if (!validAddress(rec.chain, rec.maker)) throw err("maker", JSON.stringify(rec.maker) + " is not a canonical " + rec.chain + " address");
    if (!validHash(rec.nonce)) throw err("nonce", JSON.stringify(rec.nonce) + " is not 32 non-zero lowercase hex bytes");
    if (!ADAPTER.test(rec.adapter)) throw err("adapter", JSON.stringify(rec.adapter) + " is not <name>/<major>.<minor>.<patch>");
    if (!validHash(rec.ref)) throw err("ref", JSON.stringify(rec.ref) + " is not 32 non-zero lowercase hex bytes");
    return true;
  }

  // Strict JSON input: unknown fields, duplicate keys, trailing data and non-canonical values are refused.
  function parse(text) {
    if (typeof text !== "string") throw err("signal", "record is not JSON text");
    let rec;
    try { rec = JSON.parse(text); } catch (e) { throw err("signal", "invalid JSON"); }
    const rawV = checkNoDuplicateKeys(text);
    // JSON.parse turns 1.0 and 1e0 into 1, while encoding/json refuses both for an int field.
    if (rawV !== null && rawV !== "1") throw err("v", "got " + rawV + ", this reader accepts " + VERSION);
    validate(rec);
    return rec;
  }

  // Canonical bytes are the fields in Intent order, no whitespace, and no HTML escaping.
  function canonical(rec) {
    validate(rec);
    return JSON.stringify({
      v: rec.v, chain: rec.chain, asset: rec.asset, token: rec.token, side: rec.side, price: rec.price,
      pay: rec.pay, expiry: rec.expiry, maker: rec.maker, nonce: rec.nonce, adapter: rec.adapter, ref: rec.ref,
    });
  }

  function parseExpiry(s) {
    if (typeof s !== "string") throw err("expiry", "is not time:<seconds> or height:<blocks>");
    const i = s.indexOf(":");
    if (i < 0 || i === s.length - 1) throw err("expiry", JSON.stringify(s) + " is not time:<seconds> or height:<blocks>");
    const kind = s.slice(0, i), raw = s.slice(i + 1);
    // strconv.ParseUint accepts leading zeroes; it does not accept a sign or a fraction.
    if (!/^[0-9]+$/.test(raw)) throw err("expiry", JSON.stringify(s) + " has no positive integer value");
    const value = BigInt(raw);
    if (value === 0n || value > MAX_U64) throw err("expiry", JSON.stringify(s) + " has no positive integer value");
    if (kind === "time") {
      if (value > MAX_I64) throw err("expiry", JSON.stringify(s) + " is beyond a Unix timestamp");
      return { kind: "time", value };
    }
    if (kind === "height") return { kind: "height", value };
    throw err("expiry", JSON.stringify(s) + " is not time:<seconds> or height:<blocks>");
  }

  // Rejects at now >= a time expiry, or at source height >= a height expiry. Height 0 means unknown:
  // a height-expiring record remains live until the caller checks the source chain.
  function live(rec, opts) {
    validate(rec);
    opts = opts || {};
    const now = opts.now === undefined ? BigInt(Math.floor(Date.now() / 1000)) : toInt64(opts.now);
    const height = opts.height === undefined ? 0n : toUint64(opts.height);
    const e = parseExpiry(rec.expiry);
    if (e.kind === "time" && now >= e.value) throw err("expiry", "time " + e.value + " has passed");
    if (e.kind === "height" && height !== 0n && height >= e.value) throw err("expiry", "source height " + height + " has reached " + e.value);
    return true;
  }

  function validToken(chain, s) {
    if (chain === "eth") return validUint(s, 256, true); // token id 0 is legal
    if (chain === "sol") return validAddress(chain, s);
    return false;
  }

  function validUint(s, maxBits, allowZero) {
    if (typeof s !== "string" || s === "" || (s[0] === "0" && s.length > 1) || !CANONICAL_UINT.test(s)) return false;
    const n = BigInt(s), max = (1n << BigInt(maxBits)) - 1n;
    return n <= max && (allowZero || n !== 0n);
  }

  function validHash(s) {
    return typeof s === "string" && s.length === 64 && LOWERCASE_HEX.test(s) && !/^0+$/.test(s);
  }

  function validAddress(chain, s) {
    if (typeof s !== "string") return false;
    if (chain === "eth") {
      if (s.length !== 42 || !s.startsWith("0x")) return false;
      const body = s.slice(2);
      return LOWERCASE_HEX.test(body) && !/^0+$/.test(body);
    }
    if (chain === "sol") {
      const b = decodeBase58(s);
      return !!b && b.length === 32 && b.some((x) => x !== 0);
    }
    return false;
  }

  function decodeBase58(s) {
    if (!s) return null;
    let n = 0n;
    for (const c of s) {
      const d = BASE58.indexOf(c);
      if (d < 0) return null;
      n = n * 58n + BigInt(d);
    }
    let bytes = [];
    if (n !== 0n) {
      let hex = n.toString(16);
      if (hex.length % 2) hex = "0" + hex;
      for (let i = 0; i < hex.length; i += 2) bytes.push(parseInt(hex.slice(i, i + 2), 16));
    }
    let pad = 0;
    while (pad < s.length && s[pad] === "1") pad++;
    const out = new Uint8Array(pad + bytes.length);
    for (let i = 0; i < bytes.length; i++) out[pad + i] = bytes[i];
    return out;
  }

  function toUint64(x) {
    let n;
    if (typeof x === "bigint") n = x;
    else if (typeof x === "number" && Number.isSafeInteger(x)) n = BigInt(x);
    else if (typeof x === "string" && /^[0-9]+$/.test(x)) n = BigInt(x);
    else throw err("expiry", "source height is not a uint64");
    if (n < 0n || n > MAX_U64) throw err("expiry", "source height is not a uint64");
    return n;
  }

  function toInt64(x) {
    let n;
    if (typeof x === "bigint") n = x;
    else if (typeof x === "number" && Number.isSafeInteger(x)) n = BigInt(x);
    else if (typeof x === "string" && /^-?[0-9]+$/.test(x)) n = BigInt(x);
    else throw err("expiry", "clock is not a Unix timestamp");
    if (n < -MAX_I64 - 1n || n > MAX_I64) throw err("expiry", "clock is not a Unix timestamp");
    return n;
  }

  // JSON.parse keeps the last duplicate, so two readers could see two different records from one
  // payload. Walk the text first and reject a repeated key at any depth. It only has to find keys;
  // JSON.parse above already rejected malformed syntax.
  function checkNoDuplicateKeys(text) {
    const n = text.length;
    let i = 0;
    let rootV = null;
    const fail = () => { throw err("signal", "invalid JSON"); };
    const ws = () => { while (i < n && /\s/.test(text[i])) i++; };
    const string = () => {
      const start = i;
      if (text[i] !== '"') fail();
      i++;
      while (i < n) {
        const c = text[i++];
        if (c === "\\") { i++; continue; }
        if (c === '"') { try { return JSON.parse(text.slice(start, i)); } catch (e) { fail(); } }
        if (c < " ") fail();
      }
      fail();
    };
    const primitive = () => {
      const start = i;
      while (i < n && !/[\s,\]}]/.test(text[i])) i++;
      if (i === start) fail();
    };
    const value = (root) => {
      ws();
      const c = text[i];
      if (c === "{") return object(!!root);
      if (c === "[") return array();
      if (c === '"') return string();
      primitive();
    };
    const object = (root) => {
      i++;
      ws();
      if (text[i] === "}") { i++; return; }
      const keys = new Set();
      for (;;) {
        ws();
        const key = string();
        if (keys.has(key)) throw err("signal", "duplicate key " + JSON.stringify(key));
        keys.add(key);
        ws();
        if (text[i++] !== ":") fail();
        // Capture the raw token for the root v, including a string or a container, so parse() can
        // reject a number spelling JSON.parse would otherwise normalise to 1.
        if (root && key === "v") {
          ws();
          const start = i;
          value(false);
          if (text[start] !== "{" && text[start] !== "[" && text[start] !== '"') rootV = text.slice(start, i);
        } else value(false);
        ws();
        const d = text[i++];
        if (d === "}") return;
        if (d !== ",") fail();
      }
    };
    const array = () => {
      i++;
      ws();
      if (text[i] === "]") { i++; return; }
      for (;;) {
        value();
        ws();
        const d = text[i++];
        if (d === "]") return;
        if (d !== ",") fail();
      }
    };
    value(true);
    ws();
    if (i !== n) fail();
    return rootV;
  }

  V.signal = Object.freeze({ version: VERSION, fields: FIELDS, parse, validate, canonical, parseExpiry, live });
})();
