package main

// AUD-15-02, -09, -10, -11. The factory's registry is permissionless: anyone may register any id for the registration
// fee, and a registration is a paid hint, never a certificate (contracts/factory.bas header). This bot used to sweep
// every pool_/launch_ entry and trust what the contract stored, so a contract anyone could install decided what the
// bot believed the treasury had earned:
//
//   - fd/fa are two stored integers, and ClaimFees can grow fc while sending nothing, so the journal credited units
//     the treasury never received and depositShare then moved the treasury's own balance into the pass (AUD-15-02);
//   - one entry with no fd key stopped the bot on log.Fatal for ever, the registry being append-only (AUD-15-09);
//   - market kind came from the list prefix, so a pool registered as a launch booked its LP token as the fee asset
//     and its deposit could never be funded (AUD-15-10);
//   - a market paying some other treasury grew fc for money this wallet never saw (AUD-15-11).
//
// The gate is the app's rule, applied here: the code hash must be one this build ships for the list the entry was
// registered under, current or legacy, and the treasury the market pays must be this bot's own wallet. Everything
// else is skipped and named. Nothing here trusts a registration to describe itself.

import (
	"fmt"
	"regexp"
	"strconv"

	rpc "veil/tools/simrpc"
)

// The template hashes config.js ships, per registration list. Set in main; a list whose set is empty sweeps nothing.
var (
	poolTpl   map[string]bool
	launchTpl map[string]bool
)

// templateHashes reads one template block's current and legacy hashes out of config.js. The block is matched by its
// key, so a hash in another template's block is never taken for this one.
func templateHashes(src, kind string) map[string]bool {
	out := map[string]bool{}
	m := regexp.MustCompile(kind + `:\s*\{[^}]*?sha256:\s*"([0-9a-f]{64})"(?:[^}]*?legacy:\s*\[([^\]]*)\])?`).FindStringSubmatch(src)
	if m == nil {
		return out
	}
	out[m[1]] = true
	for _, h := range regexp.MustCompile(`[0-9a-f]{64}`).FindAllString(m[2], -1) {
		out[h] = true
	}
	return out
}

// marketTemplates reads the pool and launch template hashes the app verifies against, current and legacy.
func marketTemplates(src string) (pool, launch map[string]bool) {
	return templateHashes(src, "pool"), templateHashes(src, "launch")
}

// listedMarket says whether one registry entry may be swept: its code must hash to a template this build ships for
// the list it came from, and it must pay this bot's own treasury. list names that list for the message.
func listedMarket(codeHash, treasury, want string, tpl map[string]bool, list string) error {
	if len(tpl) == 0 {
		return fmt.Errorf("config.js names no %s template, so no entry on this list can be verified", list)
	}
	if !tpl[codeHash] {
		return fmt.Errorf("its code does not hash to a shipped %s template, so what it stores says nothing about fees the treasury earned", list)
	}
	if treasury == "" {
		return fmt.Errorf("its treasury cannot be read, so whether it pays this wallet is unknown")
	}
	if treasury != want {
		return fmt.Errorf("it pays %s, not this treasury %s", short(treasury), short(want))
	}
	return nil
}

// listedMarkets turns the factory's registry into the markets this sweep may touch. Every entry is checked before any
// key of its storage is read for fees, so a registration cannot choose the assets the journal is made to owe.
func listedMarkets(reg rpc.SCState, ch chain, me string) []market {
	want := rpc.DVMAddress(me)
	if want == "" {
		fmt.Println("this wallet's address cannot be read as a DERO address: no market can be shown to pay it, so none is swept")
		return nil
	}
	var out []market
	for _, p := range registryPrefixes {
		tpl, list := poolTpl, "pool"
		if p.prefix == "launch_" {
			tpl, list = launchTpl, "launch"
		}
		for i := uint64(0); i < reg.U0(p.count); i++ {
			k := p.prefix + strconv.FormatUint(i, 10)
			if !reg.Has(k) {
				continue
			}
			if _, ok := reg.StringKeys[k].(string); !ok {
				fmt.Printf("skipping %s: its registry entry is not a contract id\n", k)
				continue
			}
			s := reg.Str(k)
			if len(s) != 64 {
				fmt.Printf("skipping %s: not a contract id (%d characters)\n", k, len(s))
				continue
			}
			if err := listedMarket(ch.CodeHash(s), ch.MarketTreasury(s), want, tpl, list); err != nil {
				fmt.Printf("skipping %s…: %v\n", s[:8], err)
				continue
			}
			// Read after the gate: Legs reads stored keys, and a contract that is not the shipped template need not
			// have them.
			q, a := ch.Legs(s, p.prefix == "launch_")
			if !isID(q) || !isID(a) {
				fmt.Printf("skipping %s…: its quote or asset id cannot be read, so its fees could not be deposited as what they are\n", s[:8])
				continue
			}
			out = append(out, market{s, q, a})
		}
	}
	return out
}

// MarketTreasury is the address a pool or launch pays its fees to, in the spelling ADDRESS_STRING writes ("" when the
// contract keeps no treasury key a 33-byte address can be read from).
func (d *deroChain) MarketTreasury(scid string) string {
	st := d.c.GetSC(scid, false)
	if !st.Has("treasury") {
		return ""
	}
	return rawAddress([]byte(st.Str("treasury")))
}
