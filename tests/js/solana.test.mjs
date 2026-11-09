import test from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { join } from "node:path";
import { ROOT } from "./harness.mjs";

const window = { Veil: { ui: { clean: (s) => String(s || "") } } };
new Function("window", readFileSync(join(ROOT, "app", "solana.js"), "utf8"))(window);
const X = window.Veil.solana._test;

test("base58 public keys round-trip canonically and reject invalid lengths", () => {
  const key = "So11111111111111111111111111111111111111112";
  assert.equal(X.encodeBase58(X.publicKey(key)), key);
  assert.throws(() => X.publicKey("1"), /32-byte/);
  assert.throws(() => X.publicKey("0".repeat(44)), /base58/);
});

test("Metaplex metadata PDA matches the published wrapped-SOL mainnet account", async () => {
  const pda = await X.metadataPda("So11111111111111111111111111111111111111112");
  assert.equal(pda, "6dM4TqWyWJsbx7obrdLcviBkTafD5E8av61zfU6jq57X");
  const ed = await X.editionPda("So11111111111111111111111111111111111111112");
  assert.equal(X.publicKey(ed).length, 32);
  assert.notEqual(ed, pda);
});

test("SPL mint parser only accepts zero-decimal supply-one mints owned by the expected program", () => {
  const owner = "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA";
  const data = new Uint8Array(82); data[44] = 0; data[45] = 1; data[36] = 1;
  const account = () => ({ owner, data: [Buffer.from(data).toString("base64"), "base64"] });
  assert.equal(X.parseMint(account(), owner), true);
  data[36] = 2;
  assert.equal(X.parseMint(account(), owner), false, "supply greater than one is not a 1/1 mint");
  data[36] = 1; data[44] = 1;
  assert.equal(X.parseMint(account(), owner), false, "decimal tokens are not NFTs");
  assert.equal(X.parseMint(account(), "TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb"), null, "owner program must match");
});

test("raw token account bytes must agree on program, mint, owner, state and exact amount", () => {
  const tokenProgram = "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA";
  const mint = "So11111111111111111111111111111111111111112", owner = "11111111111111111111111111111111";
  const bytes = new Uint8Array(165); bytes.set(X.publicKey(mint), 0); bytes.set(X.publicKey(owner), 32); bytes[64] = 1; bytes[108] = 1;
  const account = () => ({ owner: tokenProgram, data: [Buffer.from(bytes).toString("base64"), "base64"] });
  assert.equal(X.parseTokenAccount(account(), tokenProgram, owner, mint, 1n), true);
  assert.equal(X.parseTokenAccount(account(), tokenProgram, mint, mint, 1n), false, "a different raw account owner is rejected");
  assert.equal(X.parseTokenAccount(account(), tokenProgram, owner, mint, 2n), false, "a different raw amount is rejected");
  bytes[108] = 0;
  assert.equal(X.parseTokenAccount(account(), tokenProgram, owner, mint, 1n), false, "uninitialized accounts are rejected");
  assert.equal(X.parseTokenAccount(account(), "11111111111111111111111111111111", owner, mint, 1n), false, "a different token program is rejected");
});

test("metadata parser checks account key and mint, then reads bounded strings", () => {
  const mint = "So11111111111111111111111111111111111111112";
  const bytes = new Uint8Array(160);
  bytes[0] = 4; bytes.set(X.publicKey(mint), 33);
  let at = 65;
  const field = (s) => { const b = new TextEncoder().encode(s); bytes[at++] = b.length; bytes.fill(0, at, at + 3); at += 3; bytes.set(b, at); at += b.length; };
  field("DMZ Art"); field("DMZ"); field("https://metadata.invalid/nft.json");
  const parsed = X.parseMetadata(bytes, mint);
  assert.equal(parsed.name, "DMZ Art"); assert.equal(parsed.symbol, "DMZ"); assert.equal(parsed.uri, "https://metadata.invalid/nft.json");
  assert.throws(() => X.parseMetadata(bytes, "11111111111111111111111111111111"), /different mint/);
  bytes[0] = 3; assert.throws(() => X.parseMetadata(bytes, mint), /Not a Metaplex/);
});

test("RPC endpoint accepts HTTPS and loopback, keeps provider query keys, and never embeds credentials", () => {
  assert.equal(X.endpointUrl("https://rpc.example.test/mainnet"), "https://rpc.example.test/mainnet");
  assert.equal(X.endpointUrl("http://localhost:8899"), "http://localhost:8899/");
  assert.equal(X.endpointUrl("https://rpc.example.test/?api-key=secret"), "https://rpc.example.test/?api-key=secret");
  for (const url of ["http://rpc.example.test", "https://user:pass@rpc.example.test", "https://rpc.example.test/#fragment", "javascript:alert(1)", "not a url"])
    assert.equal(X.endpointUrl(url), null, url);
});
