// DMZ solana.js — read-only Solana mainnet NFT gallery. No Solana key, signer, transaction or remote media request.
// The selected RPC can report false account state: all results are explicitly labelled as that endpoint's view. RPC URLs stay in the input and request only; DMZ does not persist them.
(function () {
  const V = window.Veil, U = V.ui, { h, render, button } = U;
  const TOKEN = "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA";
  const TOKEN_2022 = "TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb";
  const META_PROGRAM = "metaqbxxUerdq28cj1RbAWkYQm3ybzjb6a8bt518x1s";
  const MAINNET_GENESIS = "5eykt4UsFv8P8NJdTREpY1vzqKqZKvdpKuc147dw2N9d";
  const PDA_MARKER = new TextEncoder().encode("ProgramDerivedAddress");
  const B58 = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz";
  const P = (1n << 255n) - 19n, MOD = (x) => ((x % P) + P) % P;
  const D = MOD(-121665n * pow(121666n, P - 2n));
  let requestId = 0;

  function pow(b, e) { let x = MOD(b), y = 1n; while (e > 0n) { if (e & 1n) y = MOD(y * x); x = MOD(x * x); e >>= 1n; } return y; }
  function decodeBase58(s) {
    if (typeof s !== "string" || !s.length || s.length > 128) throw new Error("Invalid base58 public key");
    let n = 0n;
    for (const c of s) { const d = B58.indexOf(c); if (d < 0) throw new Error("Invalid base58 public key"); n = n * 58n + BigInt(d); }
    const tail = []; while (n) { tail.push(Number(n & 255n)); n >>= 8n; }
    let zeroes = 0; while (zeroes < s.length && s[zeroes] === "1") zeroes++;
    const out = new Uint8Array(zeroes + tail.length); for (let i = 0; i < tail.length; i++) out[out.length - 1 - i] = tail[i];
    return out;
  }
  function encodeBase58(bytes) {
    let n = 0n; for (const b of bytes) n = (n << 8n) | BigInt(b);
    let out = ""; while (n) { const d = Number(n % 58n); out = B58[d] + out; n /= 58n; }
    let zeroes = 0; while (zeroes < bytes.length && bytes[zeroes] === 0) zeroes++;
    return "1".repeat(zeroes) + out;
  }
  function publicKey(s) { const b = decodeBase58(s); if (b.length !== 32 || encodeBase58(b) !== s) throw new Error("Expected a canonical 32-byte Solana public key"); return b; }
  // Compressed Edwards-Y point test. Solana PDAs are valid only when their digest is NOT a decompressible Ed25519 point.
  function isEd25519Point(bytes) {
    if (!(bytes instanceof Uint8Array) || bytes.length !== 32) return false;
    const b = bytes.slice(); const sign = b[31] >>> 7; b[31] &= 127;
    let y = 0n; for (let i = 31; i >= 0; i--) y = (y << 8n) | BigInt(b[i]);
    if (y >= P) return false;
    const y2 = MOD(y * y), u = MOD(y2 - 1n), v = MOD(D * y2 + 1n); if (v === 0n) return false;
    const x2 = MOD(u * pow(v, P - 2n)); let x = pow(x2, (P + 3n) / 8n);
    if (MOD(x * x - x2) !== 0n) x = MOD(x * pow(2n, (P - 1n) / 4n));
    if (MOD(x * x - x2) !== 0n || (x === 0n && sign === 1)) return false;
    return true;
  }
  async function createProgramAddress(seeds, program) {
    if (seeds.length > 16 || seeds.some((s) => !(s instanceof Uint8Array) || s.length > 32)) throw new Error("Invalid Solana PDA seeds");
    const chunks = [...seeds, publicKey(program), PDA_MARKER], size = chunks.reduce((n, x) => n + x.length, 0), all = new Uint8Array(size);
    let at = 0; for (const c of chunks) { all.set(c, at); at += c.length; }
    const digest = new Uint8Array(await crypto.subtle.digest("SHA-256", all));
    if (isEd25519Point(digest)) return null;
    return encodeBase58(digest);
  }
  async function findProgramAddress(seeds, program) {
    for (let bump = 255; bump >= 0; bump--) { const pda = await createProgramAddress([...seeds, new Uint8Array([bump])], program); if (pda) return pda; }
    throw new Error("No valid Solana program address");
  }
  const metadataPda = (mint) => findProgramAddress([new TextEncoder().encode("metadata"), publicKey(META_PROGRAM), publicKey(mint)], META_PROGRAM);
  const editionPda = (mint) => findProgramAddress([new TextEncoder().encode("metadata"), publicKey(META_PROGRAM), publicKey(mint), new TextEncoder().encode("edition")], META_PROGRAM);

  const u32 = (b, i) => { if (i + 4 > b.length) throw new Error("Truncated Metaplex metadata"); return (b[i] | b[i + 1] << 8 | b[i + 2] << 16 | b[i + 3] * 16777216) >>> 0; };
  function readString(b, i, max) {
    const n = u32(b, i); if (n > max || i + 4 + n > b.length) throw new Error("Invalid Metaplex metadata string");
    return [new TextDecoder("utf-8", { fatal: false }).decode(b.subarray(i + 4, i + 4 + n)).replace(/[\u0000-\u001f\u007f-\u009f\u061c\u180e\u200b-\u200f\u202a-\u202e\u2060-\u2069\ufeff]/g, "").trim(), i + 4 + n];
  }
  function parseMetadata(bytes, expectedMint) {
    if (!(bytes instanceof Uint8Array) || bytes.length < 65 || bytes[0] !== 4) throw new Error("Not a Metaplex Metadata account");
    const mint = encodeBase58(bytes.subarray(33, 65)); if (mint !== expectedMint) throw new Error("Metaplex account names a different mint");
    let i = 65; const [name, n] = readString(bytes, i, 256); i = n;
    const [symbol, s] = readString(bytes, i, 64); i = s;
    const [uri] = readString(bytes, i, 1024);
    return { name: U.clean(name, 80), symbol: U.clean(symbol, 24), uri: U.clean(uri, 1024) };
  }
  function accountData(account) {
    const v = account && account.data;
    if (!v || !Array.isArray(v) || typeof v[0] !== "string" || v[1] !== "base64") throw new Error("RPC returned an unexpected account encoding");
    try { const s = atob(v[0]), out = new Uint8Array(s.length); for (let i = 0; i < s.length; i++) out[i] = s.charCodeAt(i); return out; }
    catch (_) { throw new Error("RPC returned invalid base64 account data"); }
  }
  function parseMint(account, expectedProgram) {
    if (!account || account.owner !== expectedProgram) return null;
    const b = accountData(account); if (b.length < 82) return null;
    let supply = 0n; for (let i = 43; i >= 36; i--) supply = (supply << 8n) | BigInt(b[i]);
    return b[44] === 0 && b[45] === 1 && supply === 1n;
  }
  function parseTokenAccount(account, expectedProgram, expectedOwner, expectedMint, expectedAmount) {
    if (!account || account.owner !== expectedProgram) return false;
    const b = accountData(account); if (b.length < 165 || encodeBase58(b.subarray(0, 32)) !== expectedMint || encodeBase58(b.subarray(32, 64)) !== expectedOwner || (b[108] !== 1 && b[108] !== 2)) return false;
    let amount = 0n; for (let i = 71; i >= 64; i--) amount = (amount << 8n) | BigInt(b[i]);
    return amount === expectedAmount;
  }
  function endpointUrl(s) {
    try {
      const u = new URL(String(s || "").trim());
      if (u.username || u.password || u.hash || (u.protocol !== "https:" && !(u.protocol === "http:" && ["localhost", "127.0.0.1", "[::1]"].includes(u.hostname)))) return null;
      return u.href;
    } catch (_) { return null; }
  }
  async function rpc(url, method, params, signal) {
    let res;
    try { res = await fetch(url, { method: "POST", mode: "cors", cache: "no-store", credentials: "omit", referrerPolicy: "no-referrer", signal,
      headers: { "Content-Type": "application/json" }, body: JSON.stringify({ jsonrpc: "2.0", id: ++requestId, method, params }) }); }
    catch (_) { throw new Error("Solana RPC request failed or was blocked by CORS. Check the endpoint and try again."); }
    if (!res.ok) throw new Error(res.status === 429 ? "Solana RPC rate limit reached (429); wait or use your own RPC endpoint." : "Solana RPC returned HTTP " + res.status + ".");
    let body; try { body = await res.json(); } catch (_) { throw new Error("Solana RPC returned invalid JSON."); }
    if (!body || body.error || !Object.prototype.hasOwnProperty.call(body, "result")) throw new Error("Solana RPC returned an error or malformed response.");
    return body.result;
  }
  const chunks = (xs, n) => Array.from({ length: Math.ceil(xs.length / n) }, (_, i) => xs.slice(i * n, (i + 1) * n));

  async function ownedNfts(url, owner, signal, progress) {
    publicKey(owner); const programs = [TOKEN, TOKEN_2022], byMint = new Map(); let scannedAccounts = 0;
    for (const programId of programs) {
      progress("Reading SPL token accounts…");
      const result = await rpc(url, "getTokenAccountsByOwner", [owner, { programId }, { encoding: "jsonParsed", commitment: "finalized" }], signal);
      if (!result || !result.context || !Array.isArray(result.value)) throw new Error("Solana RPC returned an incomplete token-account response.");
      scannedAccounts += result.value.length;
      if (scannedAccounts > 5000) throw new Error("This wallet has over 5,000 token accounts; this bounded gallery will not scan them.");
      for (const item of result.value) {
        const a = item && item.account, parsed = a && a.data && a.data.parsed, info = parsed && parsed.info, t = info && info.tokenAmount;
        if (!a || a.owner !== programId || !parsed || !["spl-token", "spl-token-2022"].includes(a.data.program) || !info || info.owner !== owner ||
          typeof item.pubkey !== "string" || typeof info.mint !== "string" || !/^\d{1,20}$/.test(String(t && t.amount)) || !Number.isInteger(t.decimals) || t.decimals !== 0 || !["initialized", "frozen"].includes(info.state)) continue;
        try { publicKey(info.mint); publicKey(item.pubkey); } catch (_) { continue; }
        const units = BigInt(t.amount); if (units === 0n) continue;
        const old = byMint.get(info.mint) || { mint: info.mint, programId, amount: 0n, tokenAccounts: [] };
        if (old.programId !== programId) continue;
        old.amount += units; old.tokenAccounts.push({ address: item.pubkey, amount: units }); byMint.set(info.mint, old);
      }
    }
    const candidates = [...byMint.values()].filter((x) => x.amount === 1n);
    if (candidates.length > 100) throw new Error("This wallet has over 100 possible NFTs; the bounded gallery refuses to continue rather than freeze the page.");
    if (!candidates.length) return { slot: 0, items: [] };

    progress("Deriving Metaplex accounts…");
    for (const batch of chunks(candidates, 32)) await Promise.all(batch.map(async (x) => { x.metadata = await metadataPda(x.mint); x.edition = await editionPda(x.mint); }));
    const addresses = candidates.flatMap((x) => [...x.tokenAccounts.map((a) => a.address), x.mint, x.metadata, x.edition]), accountMap = new Map();
    let firstSlot = Infinity, lastSlot = 0;
    for (const part of chunks(addresses, 100)) {
      progress("Checking mint supply and Metaplex editions…");
      const result = await rpc(url, "getMultipleAccounts", [part, { encoding: "base64", commitment: "finalized" }], signal);
      if (!result || !result.context || !Array.isArray(result.value) || result.value.length !== part.length || !Number.isSafeInteger(result.context.slot)) throw new Error("Solana RPC returned an incomplete account batch.");
      firstSlot = Math.min(firstSlot, result.context.slot); lastSlot = Math.max(lastSlot, result.context.slot);
      part.forEach((key, i) => accountMap.set(key, result.value[i]));
    }
    const slotResult = { first: firstSlot, last: lastSlot };
    if (!Number.isSafeInteger(slotResult.first) || !Number.isSafeInteger(slotResult.last)) throw new Error("Solana RPC returned an invalid finalized slot.");
    const items = [];
    for (const x of candidates) {
      // Verify token-account ownership from raw SPL bytes as well as the RPC's parsed getTokenAccountsByOwner result.
      // Base Token and Token-2022 accounts share the 165-byte prefix: mint[0:32], owner[32:64], amount[64:72], state[108].
      let rawAmount = 0n, tokenAccountsOK = x.tokenAccounts.length > 0;
      for (const t of x.tokenAccounts) {
        const account = accountMap.get(t.address); let valid = false;
        try { valid = parseTokenAccount(account, x.programId, owner, x.mint, t.amount); } catch (_) {}
        if (!valid) { tokenAccountsOK = false; break; } rawAmount += t.amount;
      }
      const mintAccount = accountMap.get(x.mint), metaAccount = accountMap.get(x.metadata), edAccount = accountMap.get(x.edition);
      if (!tokenAccountsOK || rawAmount !== 1n || !parseMint(mintAccount, x.programId) || !metaAccount || metaAccount.owner !== META_PROGRAM || !edAccount || edAccount.owner !== META_PROGRAM) continue;
      let meta, ed;
      try { meta = parseMetadata(accountData(metaAccount), x.mint); ed = accountData(edAccount); } catch (_) { continue; }
      // Metaplex Key::Edition=1, MasterEditionV1=2, MasterEditionV2=6. A PDA alone is not proof of an NFT.
      if (![1, 2, 6].includes(ed[0])) continue;
      items.push({ mint: x.mint, metadata: meta, kind: ed[0] === 1 ? "edition" : "master edition" });
    }
    return { slot: slotResult, items };
  }

  function section() {
    const panel = h("section", { class: "card mt", "aria-labelledby": "solanaNftTitle" });
    const address = h("input", { autocomplete: "off", autocapitalize: "off", spellcheck: "false", "aria-label": "Solana wallet public address", placeholder: "Paste a Solana public address" });
    const endpoint = h("input", { type: "password", inputmode: "url", autocomplete: "off", autocapitalize: "off", spellcheck: "false", "aria-label": "Solana mainnet RPC endpoint", placeholder: "https://your-cors-enabled-solana-rpc" });
    const status = h("div", { class: "note mt8", role: "status", "aria-live": "polite" });
    const results = h("div"); let active = null, read = false;
    const scan = async () => {
      if (active) active.abort(); active = null; read = true; render(results);
      const url = endpointUrl(endpoint.value), who = address.value.trim();
      if (!url) { render(status, h("div", { class: "errbox" }, "Use an HTTPS Solana RPC endpoint (or localhost for development). Embedded credentials and URL fragments are refused; provider API-key query parameters are kept in memory only.")); return; }
      try { publicKey(who); } catch (_) { render(status, h("div", { class: "errbox" }, "Enter a valid 32-byte Solana public address in base58.")); return; }
      if (!globalThis.crypto || !crypto.subtle) { render(status, h("div", { class: "errbox" }, "This browser does not expose Web Crypto on this page, so Solana account addresses cannot be derived safely.")); return; }
      const controller = new AbortController(); active = controller; const timer = setTimeout(() => controller.abort(), 60000);
      render(results); render(status, "Checking cluster and reading finalized account state from the selected RPC…");
      try {
        const genesis = await rpc(url, "getGenesisHash", [], controller.signal);
        if (genesis !== MAINNET_GENESIS) throw new Error("This endpoint is not Solana mainnet-beta (genesis hash mismatch); no NFTs were loaded.");
        const result = await ownedNfts(url, who, controller.signal, (t) => { if (active === controller) render(status, t); });
        if (active !== controller) return;
        render(status, result.items.length
          ? "RPC-reported ownership · Solana mainnet-beta · finalized read slots " + result.slot.first + (result.slot.last === result.slot.first ? "" : "–" + result.slot.last) + ". Account checks passed against this endpoint; this is not an independent proof of chain state."
          : "No standard SPL / Token-2022 NFTs with Metaplex metadata and edition accounts were reported for this address at finalized commitment. Compressed NFTs and non-Metaplex assets are not scanned.");
        render(results, result.items.map((x) => h("article", { class: "card mt" },
          h("div", { class: "between" }, h("b", null, x.metadata.name || "Unnamed Solana NFT"), h("span", { class: "pill dim" }, x.kind)),
          x.metadata.symbol ? h("div", { class: "sub mt8" }, x.metadata.symbol) : null,
          h("div", { class: "note mt8" }, "Mint · " + x.mint),
          x.metadata.uri ? h("div", { class: "note mt8" }, "Metadata URI (not fetched) · " + x.metadata.uri) : null)));
      } catch (e) {
        if (active !== controller) return;
        const message = controller.signal.aborted ? "Solana RPC read timed out or was cancelled. Nothing was signed." : U.clean(e && e.message || "Solana read failed", 240);
        render(status, h("div", { class: "errbox" }, message)); render(results);
      } finally { clearTimeout(timer); if (active === controller) active = null; }
    };
    const loadButton = button("Read NFTs", "small primary", scan);
    render(panel,
      h("div", { class: "between" }, h("h2", { id: "solanaNftTitle" }, "Solana · read-only gallery"), h("span", { class: "pill dim" }, "RPC-reported")),
      h("div", { class: "note mt8" }, "Paste a Solana public address and a CORS-enabled mainnet RPC endpoint (including an API-key query string if your provider requires one). The endpoint stays in this field for this page only and is not stored. DMZ checks the endpoint's genesis hash, finalized SPL token accounts, mint supply, and the canonical Metaplex metadata and edition accounts. The endpoint learns which address you query and can lie; these checks are against that endpoint, not an independent light client. No keys, signatures, transactions, or images are requested. Compressed NFTs and non-Metaplex assets are not included."),
      h("div", { class: "field mt" }, "Solana wallet address · not your DERO address", address),
      h("div", { class: "field mt8" }, "Solana mainnet RPC endpoint · choose one you trust", endpoint),
      h("div", { class: "row mt" }, loadButton), status, results);
    const changed = () => { if (active) active.abort(); active = null; render(results); if (read) render(status, "Inputs changed. Read again to query the selected address and endpoint."); };
    address.addEventListener("input", changed); endpoint.addEventListener("input", changed);
    address.addEventListener("keydown", (e) => { if (e.key === "Enter") scan(); });
    endpoint.addEventListener("keydown", (e) => { if (e.key === "Enter") scan(); });
    return panel;
  }

  V.solana = { section, _test: { decodeBase58, encodeBase58, publicKey, isEd25519Point, findProgramAddress, metadataPda, editionPda, parseMetadata, parseMint, parseTokenAccount, endpointUrl } };
})();
